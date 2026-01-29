package storage

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/fazt-sh/fazt/internal/debug"
	"github.com/fazt-sh/fazt/internal/timeout"
)

// InjectStorageNamespace adds fazt.storage.* to a Goja VM.
// The context is used for all storage operations to respect request timeouts.
// The budget parameter enables admission control and per-operation timeouts.
func InjectStorageNamespace(vm *goja.Runtime, storage *Storage, appID string, ctx context.Context, budget *timeout.Budget) error {
	// Get or create fazt object
	faztVal := vm.Get("fazt")
	var fazt *goja.Object
	if faztVal == nil || goja.IsUndefined(faztVal) {
		fazt = vm.NewObject()
		vm.Set("fazt", fazt)
	} else {
		fazt = faztVal.ToObject(vm)
	}

	storageObj := vm.NewObject()

	// fazt.storage.kv
	kvObj := vm.NewObject()
	kvObj.Set("set", makeKVSet(vm, storage.KV, appID, ctx, budget))
	kvObj.Set("get", makeKVGet(vm, storage.KV, appID, ctx, budget))
	kvObj.Set("delete", makeKVDelete(vm, storage.KV, appID, ctx, budget))
	kvObj.Set("list", makeKVList(vm, storage.KV, appID, ctx, budget))
	storageObj.Set("kv", kvObj)

	// fazt.storage.ds
	dsObj := vm.NewObject()
	dsObj.Set("insert", makeDSInsert(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("find", makeDSFind(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("findOne", makeDSFindOne(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("update", makeDSUpdate(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("delete", makeDSDelete(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("count", makeDSCount(vm, storage.Docs, appID, ctx, budget))
	dsObj.Set("deleteOldest", makeDSDeleteOldest(vm, storage.Docs, appID, ctx, budget))
	storageObj.Set("ds", dsObj)

	// fazt.storage.s3
	s3Obj := vm.NewObject()
	s3Obj.Set("put", makeS3Put(vm, storage.Blobs, appID, ctx, budget))
	s3Obj.Set("get", makeS3Get(vm, storage.Blobs, appID, ctx, budget))
	s3Obj.Set("delete", makeS3Delete(vm, storage.Blobs, appID, ctx, budget))
	s3Obj.Set("list", makeS3List(vm, storage.Blobs, appID, ctx, budget))
	storageObj.Set("s3", s3Obj)

	fazt.Set("storage", storageObj)
	return nil
}

// getOpContext creates a scoped context for a storage operation.
// If budget is nil, returns the parent context unchanged.
// If budget has insufficient time, returns an error.
func getOpContext(vm *goja.Runtime, parent context.Context, budget *timeout.Budget) (context.Context, func(), error) {
	if budget == nil {
		return parent, func() {}, nil
	}

	if !budget.CanStartOperation() {
		return nil, nil, timeout.ErrInsufficientTime
	}

	ctx, cancel, err := budget.StorageContext(parent)
	if err != nil {
		return nil, nil, err
	}
	return ctx, cancel, nil
}

// KV bindings

func makeKVSet(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		if err := kv.Set(opCtx, appID, key, value, ttl); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeKVGet(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		value, err := kv.Get(opCtx, appID, key)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		if value == nil {
			return goja.Undefined()
		}

		return vm.ToValue(value)
	}
}

func makeKVDelete(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		if err := kv.Delete(opCtx, appID, key); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeKVList(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		entries, err := kv.List(opCtx, appID, prefix)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		// Convert to JS array of objects
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

// Document store bindings

func makeDSInsert(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		// Warn if trying to set reserved field
		if _, hasID := doc["id"]; hasID {
			debug.Warn("storage", "ds.insert: 'id' in doc will be used as document ID")
		}

		id, err := ds.Insert(opCtx, appID, collection, doc)
		debug.StorageOp("insert", appID, collection, doc, 1, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(id)
	}
}

func makeDSFind(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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
			queryVal := call.Argument(1).Export()
			if q, ok := queryVal.(map[string]interface{}); ok {
				query = q
			}
		}

		// Parse options (3rd argument): { limit, offset, order }
		var opts *FindOptions
		if len(call.Arguments) >= 3 && !goja.IsUndefined(call.Argument(2)) && !goja.IsNull(call.Argument(2)) {
			optsVal := call.Argument(2).Export()
			if o, ok := optsVal.(map[string]interface{}); ok {
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

		// Use the extended interface if available
		var docs []Document
		if sqlDS, ok := ds.(*SQLDocStore); ok && opts != nil {
			docs, err = sqlDS.FindWithOptions(opCtx, appID, collection, query, opts)
		} else {
			docs, err = ds.Find(opCtx, appID, collection, query)
		}
		debug.StorageOp("find", appID, collection, query, int64(len(docs)), time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		// Convert to JS array
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

func makeDSFindOne(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		// Accept either string ID (legacy) or query object
		var query map[string]interface{}
		arg1 := call.Argument(1)
		exported := arg1.Export()

		switch v := exported.(type) {
		case string:
			// Legacy: ds.findOne('col', 'id-string')
			query = map[string]interface{}{"id": v}
		case map[string]interface{}:
			// New: ds.findOne('col', { id: 'x', session: 'y' })
			query = v
		default:
			panic(vm.NewGoError(fmt.Errorf("ds.findOne requires query object or string ID, got %T", exported)))
		}

		docs, err := ds.Find(opCtx, appID, collection, query)
		rows := int64(0)
		if len(docs) > 0 {
			rows = 1
		}
		debug.StorageOp("findOne", appID, collection, query, rows, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		if len(docs) == 0 {
			return goja.Null()
		}

		doc := docs[0]
		obj := doc.Data
		obj["id"] = doc.ID
		obj["_createdAt"] = doc.CreatedAt.UnixMilli()
		obj["_updatedAt"] = doc.UpdatedAt.UnixMilli()

		return vm.ToValue(obj)
	}
}

func makeDSUpdate(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		queryVal := call.Argument(1).Export()
		query, ok := queryVal.(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.update requires a query object")))
		}

		changesVal := call.Argument(2).Export()
		changes, ok := changesVal.(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.update requires a changes object")))
		}

		count, err := ds.Update(opCtx, appID, collection, query, changes)
		debug.StorageOp("update", appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

func makeDSDelete(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		queryVal := call.Argument(1).Export()
		query, ok := queryVal.(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.delete requires a query object")))
		}

		count, err := ds.Delete(opCtx, appID, collection, query)
		debug.StorageOp("delete", appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

func makeDSCount(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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
			queryVal := call.Argument(1).Export()
			if q, ok := queryVal.(map[string]interface{}); ok {
				query = q
			}
		}

		// Use SQLDocStore.Count if available
		var count int64
		if sqlDS, ok := ds.(*SQLDocStore); ok {
			count, err = sqlDS.Count(opCtx, appID, collection, query)
		} else {
			// Fallback: count via Find (less efficient)
			docs, findErr := ds.Find(opCtx, appID, collection, query)
			if findErr != nil {
				err = findErr
			} else {
				count = int64(len(docs))
			}
		}
		debug.StorageOp("count", appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

func makeDSDeleteOldest(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("ds.deleteOldest requires collection and keepCount")))
		}

		opCtx, cancel, err := getOpContext(vm, ctx, budget)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer cancel()

		collection := call.Argument(0).String()
		keepCount := int(call.Argument(1).ToInteger())

		// Only SQLDocStore supports this operation
		sqlDS, ok := ds.(*SQLDocStore)
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.deleteOldest requires SQLDocStore")))
		}

		count, err := sqlDS.DeleteOldest(opCtx, appID, collection, keepCount)
		debug.StorageOp("deleteOldest", appID, collection, map[string]interface{}{"keepCount": keepCount}, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

// Blob store bindings

func makeS3Put(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		// Handle data - can be string or ArrayBuffer
		var data []byte
		dataArg := call.Argument(1)
		exported := dataArg.Export()

		switch v := exported.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		case goja.ArrayBuffer:
			data = v.Bytes()
		default:
			// Try to get as base64 string
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

		if err := blobs.Put(opCtx, appID, path, data, mimeType); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeS3Get(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		blob, err := blobs.Get(opCtx, appID, path)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		if blob == nil {
			return goja.Null()
		}

		// Return object with data as base64 and metadata
		result := map[string]interface{}{
			"data": base64.StdEncoding.EncodeToString(blob.Data),
			"mime": blob.MimeType,
			"size": blob.Size,
			"hash": blob.Hash,
		}

		return vm.ToValue(result)
	}
}

func makeS3Delete(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		if err := blobs.Delete(opCtx, appID, path); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeS3List(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context, budget *timeout.Budget) func(goja.FunctionCall) goja.Value {
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

		items, err := blobs.List(opCtx, appID, prefix)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		// Convert to JS array
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
