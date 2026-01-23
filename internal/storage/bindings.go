package storage

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/fazt-sh/fazt/internal/debug"
)

// InjectStorageNamespace adds fazt.storage.* to a Goja VM.
// The context is used for all storage operations to respect request timeouts.
func InjectStorageNamespace(vm *goja.Runtime, storage *Storage, appID string, ctx context.Context) error {
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
	kvObj.Set("set", makeKVSet(vm, storage.KV, appID, ctx))
	kvObj.Set("get", makeKVGet(vm, storage.KV, appID, ctx))
	kvObj.Set("delete", makeKVDelete(vm, storage.KV, appID, ctx))
	kvObj.Set("list", makeKVList(vm, storage.KV, appID, ctx))
	storageObj.Set("kv", kvObj)

	// fazt.storage.ds
	dsObj := vm.NewObject()
	dsObj.Set("insert", makeDSInsert(vm, storage.Docs, appID, ctx))
	dsObj.Set("find", makeDSFind(vm, storage.Docs, appID, ctx))
	dsObj.Set("findOne", makeDSFindOne(vm, storage.Docs, appID, ctx))
	dsObj.Set("update", makeDSUpdate(vm, storage.Docs, appID, ctx))
	dsObj.Set("delete", makeDSDelete(vm, storage.Docs, appID, ctx))
	storageObj.Set("ds", dsObj)

	// fazt.storage.s3
	s3Obj := vm.NewObject()
	s3Obj.Set("put", makeS3Put(vm, storage.Blobs, appID, ctx))
	s3Obj.Set("get", makeS3Get(vm, storage.Blobs, appID, ctx))
	s3Obj.Set("delete", makeS3Delete(vm, storage.Blobs, appID, ctx))
	s3Obj.Set("list", makeS3List(vm, storage.Blobs, appID, ctx))
	storageObj.Set("s3", s3Obj)

	fazt.Set("storage", storageObj)
	return nil
}

// KV bindings

func makeKVSet(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("kv.set requires key and value")))
		}

		key := call.Argument(0).String()
		value := call.Argument(1).Export()

		var ttl *time.Duration
		if len(call.Arguments) >= 3 && !goja.IsUndefined(call.Argument(2)) && !goja.IsNull(call.Argument(2)) {
			ms := call.Argument(2).ToInteger()
			d := time.Duration(ms) * time.Millisecond
			ttl = &d
		}

		if err := kv.Set(ctx, appID, key, value, ttl); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeKVGet(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("kv.get requires a key")))
		}

		key := call.Argument(0).String()

		value, err := kv.Get(ctx, appID, key)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		if value == nil {
			return goja.Undefined()
		}

		return vm.ToValue(value)
	}
}

func makeKVDelete(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("kv.delete requires a key")))
		}

		key := call.Argument(0).String()

		if err := kv.Delete(ctx, appID, key); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeKVList(vm *goja.Runtime, kv KVStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		prefix := ""
		if len(call.Arguments) >= 1 && !goja.IsUndefined(call.Argument(0)) {
			prefix = call.Argument(0).String()
		}

		entries, err := kv.List(ctx, appID, prefix)
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

func makeDSInsert(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("ds.insert requires collection and document")))
		}

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

		id, err := ds.Insert(ctx, appID, collection, doc)
		debug.StorageOp("insert", appID, collection, doc, 1, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(id)
	}
}

func makeDSFind(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("ds.find requires collection")))
		}

		collection := call.Argument(0).String()

		query := make(map[string]interface{})
		if len(call.Arguments) >= 2 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			queryVal := call.Argument(1).Export()
			if q, ok := queryVal.(map[string]interface{}); ok {
				query = q
			}
		}

		docs, err := ds.Find(ctx, appID, collection, query)
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

func makeDSFindOne(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("ds.findOne requires collection and query")))
		}

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

		docs, err := ds.Find(ctx, appID, collection, query)
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

func makeDSUpdate(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 3 {
			panic(vm.NewGoError(fmt.Errorf("ds.update requires collection, query, and changes")))
		}

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

		count, err := ds.Update(ctx, appID, collection, query, changes)
		debug.StorageOp("update", appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

func makeDSDelete(vm *goja.Runtime, ds DocStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		start := time.Now()
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("ds.delete requires collection and query")))
		}

		collection := call.Argument(0).String()

		queryVal := call.Argument(1).Export()
		query, ok := queryVal.(map[string]interface{})
		if !ok {
			panic(vm.NewGoError(fmt.Errorf("ds.delete requires a query object")))
		}

		count, err := ds.Delete(ctx, appID, collection, query)
		debug.StorageOp("delete", appID, collection, query, count, time.Since(start))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(count)
	}
}

// Blob store bindings

func makeS3Put(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("s3.put requires path and data")))
		}

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

		if err := blobs.Put(ctx, appID, path, data, mimeType); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeS3Get(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("s3.get requires a path")))
		}

		path := call.Argument(0).String()

		blob, err := blobs.Get(ctx, appID, path)
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

func makeS3Delete(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewGoError(fmt.Errorf("s3.delete requires a path")))
		}

		path := call.Argument(0).String()

		if err := blobs.Delete(ctx, appID, path); err != nil {
			panic(vm.NewGoError(err))
		}

		return goja.Undefined()
	}
}

func makeS3List(vm *goja.Runtime, blobs BlobStore, appID string, ctx context.Context) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		prefix := ""
		if len(call.Arguments) >= 1 && !goja.IsUndefined(call.Argument(0)) {
			prefix = call.Argument(0).String()
		}

		items, err := blobs.List(ctx, appID, prefix)
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
