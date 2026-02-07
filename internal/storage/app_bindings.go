package storage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/fazt-sh/fazt/internal/debug"
	"github.com/fazt-sh/fazt/internal/services/media"
	"github.com/fazt-sh/fazt/internal/timeout"
)

// InjectAppNamespace adds fazt.app.* namespace to a Goja VM.
// This is the new API that replaces fazt.storage.*.
//
// Structure:
//   - fazt.app.user.ds/kv/s3 - user's private data (requires login)
//   - fazt.app.ds/kv/s3 - shared app data
func InjectAppNamespace(vm *goja.Runtime, db *sql.DB, writer *WriteQueue, appID, userID string, ctx context.Context, budget *timeout.Budget) error {
	// Get or create fazt object
	faztVal := vm.Get("fazt")
	var fazt *goja.Object
	if faztVal == nil || goja.IsUndefined(faztVal) {
		fazt = vm.NewObject()
		vm.Set("fazt", fazt)
	} else {
		fazt = faztVal.ToObject(vm)
	}

	// Get or create fazt.app object
	appVal := fazt.Get("app")
	var appObj *goja.Object
	if appVal == nil || goja.IsUndefined(appVal) {
		appObj = vm.NewObject()
		fazt.Set("app", appObj)
	} else {
		appObj = appVal.ToObject(vm)
	}

	// Create shared storage bindings: fazt.app.ds, fazt.app.kv, fazt.app.s3
	storage := New(db)

	// fazt.app.kv (shared)
	kvObj := vm.NewObject()
	kvObj.Set("set", makeKVSet(vm, storage.KV, appID, ctx, budget))
	kvObj.Set("get", makeKVGet(vm, storage.KV, appID, ctx, budget))
	kvObj.Set("delete", makeKVDelete(vm, storage.KV, appID, ctx, budget))
	kvObj.Set("list", makeKVList(vm, storage.KV, appID, ctx, budget))
	appObj.Set("kv", kvObj)

	// fazt.app.ds (shared)
	dsObj := vm.NewObject()
	dsObj.Set("insert", makeDSInsert(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("find", makeDSFind(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("findOne", makeDSFindOne(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("update", makeDSUpdate(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("delete", makeDSDelete(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("count", makeDSCount(vm, storage.Docs, appID, ctx, budget))
	appObj.Set("ds", dsObj)

	// fazt.app.s3 (shared)
	s3Obj := vm.NewObject()
	s3Obj.Set("put", makeS3PutWithMediaInvalidation(vm, storage.Blobs, appID, db, ctx, budget))
	s3Obj.Set("get", makeS3Get(vm, storage.Blobs, appID, ctx, budget))
	s3Obj.Set("delete", makeS3DeleteWithMediaInvalidation(vm, storage.Blobs, appID, db, ctx, budget))
	s3Obj.Set("list", makeS3List(vm, storage.Blobs, appID, ctx, budget))
	appObj.Set("s3", s3Obj)

	// fazt.app.media (shared)
	mediaObj := vm.NewObject()
	mediaObj.Set("serve", makeMediaServe(vm, storage.Blobs, appID, db, ctx, budget))
	mediaObj.Set("probe", makeMediaProbe(vm))
	mediaObj.Set("transcode", makeMediaTranscode(vm, storage.Blobs, appID, ctx, budget))
	appObj.Set("media", mediaObj)

	// Create user-scoped storage: fazt.app.user.*
	userObj := vm.NewObject()

	if userID != "" {
		// User is logged in - create actual user-scoped bindings
		userKV := NewUserScopedKV(db, writer, appID, userID)
		userDocs := NewUserScopedDocs(db, writer, appID, userID)
		userBlobs := NewUserScopedBlobs(db, writer, appID, userID)

		// fazt.app.user.kv
		userKVObj := vm.NewObject()
		userKVObj.Set("set", makeUserKVSet(vm, userKV, ctx, budget))
		userKVObj.Set("get", makeUserKVGet(vm, userKV, ctx, budget))
		userKVObj.Set("delete", makeUserKVDelete(vm, userKV, ctx, budget))
		userKVObj.Set("list", makeUserKVList(vm, userKV, ctx, budget))
		userObj.Set("kv", userKVObj)

		// fazt.app.user.ds
		userDSObj := vm.NewObject()
		userDSObj.Set("insert", makeUserDSInsert(vm, userDocs, ctx, budget))
		userDSObj.Set("find", makeUserDSFind(vm, userDocs, ctx, budget))
		userDSObj.Set("findOne", makeUserDSFindOne(vm, userDocs, ctx, budget))
		userDSObj.Set("update", makeUserDSUpdate(vm, userDocs, ctx, budget))
		userDSObj.Set("delete", makeUserDSDelete(vm, userDocs, ctx, budget))
		userDSObj.Set("count", makeUserDSCount(vm, userDocs, ctx, budget))
		userObj.Set("ds", userDSObj)

		// fazt.app.user.s3
		userS3Obj := vm.NewObject()
		userS3Obj.Set("put", makeUserS3PutWithMediaInvalidation(vm, userBlobs, appID, userID, db, ctx, budget))
		userS3Obj.Set("get", makeUserS3Get(vm, userBlobs, ctx, budget))
		userS3Obj.Set("delete", makeUserS3DeleteWithMediaInvalidation(vm, userBlobs, appID, userID, db, ctx, budget))
		userS3Obj.Set("list", makeUserS3List(vm, userBlobs, ctx, budget))
		userObj.Set("s3", userS3Obj)

		// fazt.app.user.media
		userMediaObj := vm.NewObject()
		userMediaObj.Set("serve", makeUserMediaServe(vm, userBlobs, appID, userID, db, ctx, budget))
		userMediaObj.Set("probe", makeMediaProbe(vm))
		userMediaObj.Set("transcode", makeUserMediaTranscode(vm, userBlobs, appID, ctx, budget))
		userObj.Set("media", userMediaObj)
	} else {
		// User not logged in - create stub bindings that throw errors
		stubFunc := func(name string) func(goja.FunctionCall) goja.Value {
			return func(call goja.FunctionCall) goja.Value {
				panic(vm.NewGoError(fmt.Errorf("fazt.app.user.%s requires login", name)))
			}
		}

		userKVObj := vm.NewObject()
		userKVObj.Set("set", stubFunc("kv.set"))
		userKVObj.Set("get", stubFunc("kv.get"))
		userKVObj.Set("delete", stubFunc("kv.delete"))
		userKVObj.Set("list", stubFunc("kv.list"))
		userObj.Set("kv", userKVObj)

		userDSObj := vm.NewObject()
		userDSObj.Set("insert", stubFunc("ds.insert"))
		userDSObj.Set("find", stubFunc("ds.find"))
		userDSObj.Set("findOne", stubFunc("ds.findOne"))
		userDSObj.Set("update", stubFunc("ds.update"))
		userDSObj.Set("delete", stubFunc("ds.delete"))
		userDSObj.Set("count", stubFunc("ds.count"))
		userObj.Set("ds", userDSObj)

		userS3Obj := vm.NewObject()
		userS3Obj.Set("put", stubFunc("s3.put"))
		userS3Obj.Set("get", stubFunc("s3.get"))
		userS3Obj.Set("delete", stubFunc("s3.delete"))
		userS3Obj.Set("list", stubFunc("s3.list"))
		userObj.Set("s3", userS3Obj)

		userMediaObj := vm.NewObject()
		userMediaObj.Set("serve", stubFunc("media.serve"))
		userObj.Set("media", userMediaObj)
	}

	appObj.Set("user", userObj)

	return nil
}

// User KV bindings

func makeUserKVSet(vm *goja.Runtime, kv *UserScopedKV, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("kv.set requires key and value")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		key := call.Argument(0).String()
		value := call.Argument(1).Export()

		var ttl *time.Duration
		if len(call.Arguments) >= 3 && !goja.IsUndefined(call.Argument(2)) && !goja.IsNull(call.Argument(2)) {
			ms := call.Argument(2).ToInteger()
			d := time.Duration(ms) * time.Millisecond
			ttl = &d
		}

		if err := kv.Set(opCtx, key, value, ttl); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeUserKVGet(vm *goja.Runtime, kv *UserScopedKV, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("kv.get requires a key")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		key := call.Argument(0).String()
		value, err := kv.Get(opCtx, key)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		if value == nil {
			return goja.Undefined()
		}
		return vm.ToValue(value)
	}
}

func makeUserKVDelete(vm *goja.Runtime, kv *UserScopedKV, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("kv.delete requires a key")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		key := call.Argument(0).String()
		if err := kv.Delete(opCtx, key); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeUserKVList(vm *goja.Runtime, kv *UserScopedKV, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		prefix := ""
		if len(call.Arguments) >= 1 && !goja.IsUndefined(call.Argument(0)) {
			prefix = call.Argument(0).String()
		}

		entries, err := kv.List(opCtx, prefix)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		result := make([]interface{}, len(entries))
		for i, entry := range entries {
			obj := map[string]interface{}{
				"key":   entry.Key,
				"value": entry.Value,
			}
			if entry.ExpiresAt != nil {
				obj["expiresAt"] = entry.ExpiresAt.UnixMilli()
			}
			result[i] = obj
		}

		return vm.ToValue(result)
	}
}

// User DS bindings

func makeUserDSInsert(vm *goja.Runtime, ds *UserScopedDocs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("ds.insert requires collection and document")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		collection := call.Argument(0).String()
		docVal := call.Argument(1).Export()

		doc, ok := docVal.(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.insert requires a document object")))
		}

		id, err := ds.Insert(opCtx, collection, doc)
		debug.StorageOp("user.insert", ds.appID, collection, doc, 1, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(id)
	}
}

func makeUserDSFind(vm *goja.Runtime, ds *UserScopedDocs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("ds.find requires collection")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		collection := call.Argument(0).String()

		query := make(map[string]interface{})
		if len(call.Arguments) >= 2 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			if q, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				query = q
			}
		}

		var opts *FindOptions
		if len(call.Arguments) >= 3 && !goja.IsUndefined(call.Argument(2)) && !goja.IsNull(call.Argument(2)) {
			if o, ok := call.Argument(2).Export().(map[string]interface{}); ok {
				opts = &FindOptions{}
				if limit, ok := o["limit"].(int64); ok {
					opts.Limit = int(limit)
				} else if limit, ok := o["limit"].(float64); ok {
					opts.Limit = int(limit)
				}
				if offset, ok := o["offset"].(int64); ok {
					opts.Offset = int(offset)
				} else if offset, ok := o["offset"].(float64); ok {
					opts.Offset = int(offset)
				}
				if order, ok := o["order"].(string); ok {
					opts.Order = order
				}
			}
		}

		docs, err := ds.FindWithOptions(opCtx, collection, query, opts)
		debug.StorageOp("user.find", ds.appID, collection, query, int64(len(docs)), time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		result := make([]interface{}, len(docs))
		for i, doc := range docs {
			obj := doc.Data
			obj["id"] = doc.ID
			obj["_createdAt"] = doc.CreatedAt.UnixMilli()
			obj["_updatedAt"] = doc.UpdatedAt.UnixMilli()
			result[i] = obj
		}

		return vm.ToValue(result)
	}
}

func makeUserDSFindOne(vm *goja.Runtime, ds *UserScopedDocs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("ds.findOne requires collection and query")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		collection := call.Argument(0).String()

		var query map[string]interface{}
		exported := call.Argument(1).Export()
		switch v := exported.(type) {
		case string:
			query = map[string]interface{}{"id": v}
		case map[string]interface{}:
			query = v
		default:
			panic(vm.NewGoError(fmt.Errorf("ds.findOne requires query object or string ID")))
		}

		doc, err := ds.FindOne(opCtx, collection, query)
		rows := int64(0)
		if doc != nil {
			rows = 1
		}
		debug.StorageOp("user.findOne", ds.appID, collection, query, rows, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		if doc == nil {
			return goja.Null()
		}

		obj := doc.Data
		obj["id"] = doc.ID
		obj["_createdAt"] = doc.CreatedAt.UnixMilli()
		obj["_updatedAt"] = doc.UpdatedAt.UnixMilli()

		return vm.ToValue(obj)
	}
}

func makeUserDSUpdate(vm *goja.Runtime, ds *UserScopedDocs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 3 {
			panic(vm.NewGoError(fmt.Errorf("ds.update requires collection, query, and changes")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		collection := call.Argument(0).String()

		query, ok := call.Argument(1).Export().(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.update requires a query object")))
		}

		changes, ok := call.Argument(2).Export().(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.update requires a changes object")))
		}

		count, err := ds.Update(opCtx, collection, query, changes)
		debug.StorageOp("user.update", ds.appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

func makeUserDSDelete(vm *goja.Runtime, ds *UserScopedDocs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("ds.delete requires collection and query")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		collection := call.Argument(0).String()

		query, ok := call.Argument(1).Export().(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.delete requires a query object")))
		}

		count, err := ds.Delete(opCtx, collection, query)
		debug.StorageOp("user.delete", ds.appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

func makeUserDSCount(vm *goja.Runtime, ds *UserScopedDocs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("ds.count requires collection")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		collection := call.Argument(0).String()

		query := make(map[string]interface{})
		if len(call.Arguments) >= 2 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			if q, ok := call.Argument(1).Export().(map[string]interface{}); ok {
				query = q
			}
		}

		count, err := ds.Count(opCtx, collection, query)
		debug.StorageOp("user.count", ds.appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

// User S3 bindings

func makeUserS3Put(vm *goja.Runtime, blobs *UserScopedBlobs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("s3.put requires path and data")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		path := call.Argument(0).String()

		var data []byte
		exported := call.Argument(1).Export()
		switch v := exported.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		case goja.ArrayBuffer:
			data = v.Bytes()
		default:
			if str, ok := exported.(string); ok {
				decoded, err := base64.StdEncoding.DecodeString(str)
				if err == nil {
					data = decoded
				} else {
					data = []byte(str)
				}
			} else {
				panic(vm.NewGoError(fmt.Errorf("s3.put data must be string or ArrayBuffer")))
			}
		}

		mimeType := "application/octet-stream"
		if len(call.Arguments) >= 3 && !goja.IsUndefined(call.Argument(2)) {
			mimeType = call.Argument(2).String()
		}

		if err := blobs.Put(opCtx, path, data, mimeType); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeUserS3Get(vm *goja.Runtime, blobs *UserScopedBlobs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("s3.get requires a path")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		path := call.Argument(0).String()
		blob, err := blobs.Get(opCtx, path)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		if blob == nil {
			return goja.Null()
		}

		result := map[string]interface{}{
			"data": base64.StdEncoding.EncodeToString(blob.Data),
			"mime": blob.MimeType,
			"size": blob.Size,
			"hash": blob.Hash,
		}

		return vm.ToValue(result)
	}
}

func makeUserS3Delete(vm *goja.Runtime, blobs *UserScopedBlobs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("s3.delete requires a path")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		path := call.Argument(0).String()
		if err := blobs.Delete(opCtx, path); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeUserS3List(vm *goja.Runtime, blobs *UserScopedBlobs, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		prefix := ""
		if len(call.Arguments) >= 1 && !goja.IsUndefined(call.Argument(0)) {
			prefix = call.Argument(0).String()
		}

		items, err := blobs.List(opCtx, prefix)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		result := make([]interface{}, len(items))
		for i, item := range items {
			result[i] = map[string]interface{}{
				"path":      item.Path,
				"mime":      item.MimeType,
				"size":      item.Size,
				"updatedAt": item.UpdatedAt.UnixMilli(),
			}
		}

		return vm.ToValue(result)
	}
}

// Media cache invalidation helpers

// makeS3PutWithMediaInvalidation wraps makeS3Put to invalidate media cache
// when an image blob is overwritten.
func makeS3PutWithMediaInvalidation(vm *goja.Runtime, blobs BlobStore, appID string, db *sql.DB, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	inner := makeS3Put(vm, blobs, appID, ctx, budget)
	return func(call goja.FunctionCall) goja.Value {
		result := inner(call)
		if len(call.Arguments) >= 3 && !goja.IsUndefined(call.Argument(2)) {
			mime := call.Argument(2).String()
			if strings.HasPrefix(mime, "image/") {
				path := call.Argument(0).String()
				media.InvalidateForPath(db, appID, path, "")
			}
		}
		return result
	}
}

// makeUserS3PutWithMediaInvalidation wraps makeUserS3Put to invalidate media cache
// when an image blob is overwritten.
func makeUserS3PutWithMediaInvalidation(vm *goja.Runtime, blobs *UserScopedBlobs, appID, userID string, db *sql.DB, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	inner := makeUserS3Put(vm, blobs, ctx, budget)
	return func(call goja.FunctionCall) goja.Value {
		result := inner(call)
		if len(call.Arguments) >= 3 && !goja.IsUndefined(call.Argument(2)) {
			mime := call.Argument(2).String()
			if strings.HasPrefix(mime, "image/") {
				path := call.Argument(0).String()
				media.InvalidateForPath(db, appID, path, userID)
			}
		}
		return result
	}
}

// makeS3DeleteWithMediaInvalidation wraps makeS3Delete to invalidate media cache
// when a blob is deleted.
func makeS3DeleteWithMediaInvalidation(vm *goja.Runtime, blobs BlobStore, appID string, db *sql.DB, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	inner := makeS3Delete(vm, blobs, appID, ctx, budget)
	return func(call goja.FunctionCall) goja.Value {
		result := inner(call)
		if len(call.Arguments) >= 1 {
			path := call.Argument(0).String()
			media.InvalidateForPath(db, appID, path, "")
		}
		return result
	}
}

// makeUserS3DeleteWithMediaInvalidation wraps makeUserS3Delete to invalidate media cache
// when a blob is deleted.
func makeUserS3DeleteWithMediaInvalidation(vm *goja.Runtime, blobs *UserScopedBlobs, appID, userID string, db *sql.DB, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	inner := makeUserS3Delete(vm, blobs, ctx, budget)
	return func(call goja.FunctionCall) goja.Value {
		result := inner(call)
		if len(call.Arguments) >= 1 {
			path := call.Argument(0).String()
			media.InvalidateForPath(db, appID, path, userID)
		}
		return result
	}
}

// Media serve bindings

// makeMediaServe creates fazt.app.media.serve(path) for shared blobs.
// Reads transform opts from HTTP query params (via context).
// On cache hit → returns cached variant. On miss → fetches original, resizes, caches.
// No transform params → returns original unchanged.
func makeMediaServe(vm *goja.Runtime, blobs BlobStore, appID string, db *sql.DB, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("media.serve requires a path")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		path := call.Argument(0).String()

		// Parse transform opts from HTTP query params
		opts := media.TransformOpts{}
		if q := media.QueryFromContext(ctx); q != nil {
			opts = media.ParseTransformQuery(q)
		}

		// No transform → serve original (prefer H.264 variant for video)
		if !opts.HasTransform() {
			blob, err := blobs.Get(opCtx, appID, path)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			if blob == nil {
				return goja.Null()
			}

			// For video: prefer H.264 variant if available
			if media.IsVideoContentType(blob.MimeType) {
				variant, err := blobs.Get(opCtx, appID, media.VariantPath(path))
				if err == nil && variant != nil {
					return vm.ToValue(map[string]interface{}{
						"data": base64.StdEncoding.EncodeToString(variant.Data),
						"mime": variant.MimeType,
						"size": variant.Size,
					})
				}
			}

			return vm.ToValue(map[string]interface{}{
				"data": base64.StdEncoding.EncodeToString(blob.Data),
				"mime": blob.MimeType,
				"size": blob.Size,
			})
		}

		// Has transform — use cache-aware processing
		cache := media.NewMediaCache(db)
		if data, mime, err := cache.Get(opCtx, appID, path, opts); err == nil && data != nil {
			return vm.ToValue(map[string]interface{}{
				"data": base64.StdEncoding.EncodeToString(data),
				"mime": mime,
				"size": len(data),
			})
		}

		// Cache miss — fetch original, process, cache
		blob, err := blobs.Get(opCtx, appID, path)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		if blob == nil {
			return goja.Null()
		}

		if !media.IsImageContentType(blob.MimeType) {
			return vm.ToValue(map[string]interface{}{
				"data": base64.StdEncoding.EncodeToString(blob.Data),
				"mime": blob.MimeType,
				"size": blob.Size,
			})
		}

		processed, mime, err := media.ProcessAndCache(opCtx, cache, appID, path, blob.Data, opts)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(map[string]interface{}{
			"data": base64.StdEncoding.EncodeToString(processed),
			"mime": mime,
			"size": len(processed),
		})
	}
}

// makeMediaTranscode creates fazt.app.media.transcode(path) for shared blobs.
// Fetches the blob, checks compatibility, and queues background transcoding if needed.
func makeMediaTranscode(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("media.transcode requires a path")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		path := call.Argument(0).String()

		blob, err := blobs.Get(opCtx, appID, path)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		if blob == nil {
			panic(vm.NewGoError(fmt.Errorf("blob not found: %s", path)))
		}

		storeFunc := func(ctx context.Context, variantPath string, data []byte, mime string) error {
			return blobs.Put(ctx, appID, variantPath, data, mime)
		}

		result := media.QueueTranscode(appID, path, blob.Data, blob.MimeType, storeFunc)
		return vm.ToValue(map[string]interface{}{
			"status": result.Status,
		})
	}
}

// makeUserMediaTranscode creates fazt.app.user.media.transcode(path) for user-scoped blobs.
func makeUserMediaTranscode(vm *goja.Runtime, blobs *UserScopedBlobs, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("media.transcode requires a path")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		path := call.Argument(0).String()

		blob, err := blobs.Get(opCtx, path)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		if blob == nil {
			panic(vm.NewGoError(fmt.Errorf("blob not found: %s", path)))
		}

		storeFunc := func(ctx context.Context, variantPath string, data []byte, mime string) error {
			return blobs.Put(ctx, variantPath, data, mime)
		}

		result := media.QueueTranscode(appID, path, blob.Data, blob.MimeType, storeFunc)
		return vm.ToValue(map[string]interface{}{
			"status": result.Status,
		})
	}
}

// makeMediaProbe creates fazt.app.media.probe(data) / fazt.app.user.media.probe(data).
// Accepts an ArrayBuffer of video data and returns codec/dimension/duration info.
func makeMediaProbe(vm *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("media.probe requires video data")))
		}

		var data []byte
		arg := call.Argument(0).Export()
		switch v := arg.(type) {
		case goja.ArrayBuffer:
			data = v.Bytes()
		default:
			if ab, ok := call.Argument(0).Export().(goja.ArrayBuffer); ok {
				data = ab.Bytes()
			} else {
				panic(vm.NewGoError(fmt.Errorf("media.probe requires ArrayBuffer, got %T", arg)))
			}
		}

		info, err := media.ProbeVideo(data)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(map[string]interface{}{
			"container":  info.Container,
			"videoCodec": info.VideoCodec,
			"audioCodec": info.AudioCodec,
			"width":      info.Width,
			"height":     info.Height,
			"duration":   info.Duration,
			"compatible": info.Compatible,
		})
	}
}

// makeUserMediaServe creates fazt.app.user.media.serve(path) for user-scoped blobs.
// Same as makeMediaServe but uses user-scoped storage and cache.
func makeUserMediaServe(vm *goja.Runtime, blobs *UserScopedBlobs, appID, userID string, db *sql.DB, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("media.serve requires a path")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		path := call.Argument(0).String()

		// Parse transform opts from HTTP query params
		opts := media.TransformOpts{}
		if q := media.QueryFromContext(ctx); q != nil {
			opts = media.ParseTransformQuery(q)
		}

		// No transform → serve original (prefer H.264 variant for video)
		if !opts.HasTransform() {
			blob, err := blobs.Get(opCtx, path)
			if err != nil {
				panic(vm.NewGoError(err))
			}
			if blob == nil {
				return goja.Null()
			}

			// For video: prefer H.264 variant if available
			if media.IsVideoContentType(blob.MimeType) {
				variant, err := blobs.Get(opCtx, media.VariantPath(path))
				if err == nil && variant != nil {
					return vm.ToValue(map[string]interface{}{
						"data": base64.StdEncoding.EncodeToString(variant.Data),
						"mime": variant.MimeType,
						"size": variant.Size,
					})
				}
			}

			return vm.ToValue(map[string]interface{}{
				"data": base64.StdEncoding.EncodeToString(blob.Data),
				"mime": blob.MimeType,
				"size": blob.Size,
			})
		}

		// Has transform — use user-scoped cache
		cache := media.NewUserMediaCache(db, userID)
		if data, mime, err := cache.Get(opCtx, appID, path, opts); err == nil && data != nil {
			return vm.ToValue(map[string]interface{}{
				"data": base64.StdEncoding.EncodeToString(data),
				"mime": mime,
				"size": len(data),
			})
		}

		// Cache miss — fetch original, process, cache
		blob, err := blobs.Get(opCtx, path)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		if blob == nil {
			return goja.Null()
		}

		if !media.IsImageContentType(blob.MimeType) {
			return vm.ToValue(map[string]interface{}{
				"data": base64.StdEncoding.EncodeToString(blob.Data),
				"mime": blob.MimeType,
				"size": blob.Size,
			})
		}

		processed, mime, err := media.ProcessAndCache(opCtx, cache, appID, path, blob.Data, opts)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(map[string]interface{}{
			"data": base64.StdEncoding.EncodeToString(processed),
			"mime": mime,
			"size": len(processed),
		})
	}
}
