package image

import (
	"fmt"

	"github.com/dop251/goja"
)

// InjectImageNamespace adds fazt.image.resize() and fazt.image.thumbnail() to a Goja VM.
// Must be called after the fazt object already exists on the VM.
func InjectImageNamespace(vm *goja.Runtime) error {
	imageObj := vm.NewObject()

	// fazt.image.resize(arrayBuffer, opts) → { data, width, height, size }
	imageObj.Set("resize", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("fazt.image.resize requires (data, options)")))
		}

		data := extractArrayBuffer(vm, call.Argument(0))
		if data == nil {
			panic(vm.NewGoError(fmt.Errorf("fazt.image.resize: first argument must be an ArrayBuffer")))
		}

		opts := parseResizeOpts(vm, call.Argument(1))

		result, err := Resize(data, opts)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return resultToJS(vm, result)
	})

	// fazt.image.thumbnail(arrayBuffer, size) → { data, width, height, size }
	imageObj.Set("thumbnail", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewGoError(fmt.Errorf("fazt.image.thumbnail requires (data, size)")))
		}

		data := extractArrayBuffer(vm, call.Argument(0))
		if data == nil {
			panic(vm.NewGoError(fmt.Errorf("fazt.image.thumbnail: first argument must be an ArrayBuffer")))
		}

		size := int(call.Argument(1).ToInteger())
		if size <= 0 {
			panic(vm.NewGoError(fmt.Errorf("fazt.image.thumbnail: size must be > 0")))
		}

		result, err := Thumbnail(data, size)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return resultToJS(vm, result)
	})

	// Attach to fazt namespace
	faztVal := vm.Get("fazt")
	if faztVal == nil || goja.IsUndefined(faztVal) {
		return fmt.Errorf("fazt object not found on VM")
	}
	fazt := faztVal.ToObject(vm)
	fazt.Set("image", imageObj)
	return nil
}

// extractArrayBuffer gets raw bytes from a JS ArrayBuffer value.
func extractArrayBuffer(vm *goja.Runtime, val goja.Value) []byte {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil
	}

	exported := val.Export()

	// Direct ArrayBuffer
	if buf, ok := exported.(goja.ArrayBuffer); ok {
		return buf.Bytes()
	}

	// []byte (some paths export as byte slice)
	if buf, ok := exported.([]byte); ok {
		return buf
	}

	return nil
}

// parseResizeOpts extracts ResizeOpts from a JS object.
func parseResizeOpts(vm *goja.Runtime, val goja.Value) ResizeOpts {
	opts := ResizeOpts{}
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return opts
	}

	obj := val.ToObject(vm)

	if v := obj.Get("width"); v != nil && !goja.IsUndefined(v) {
		opts.Width = int(v.ToInteger())
	}
	if v := obj.Get("height"); v != nil && !goja.IsUndefined(v) {
		opts.Height = int(v.ToInteger())
	}
	if v := obj.Get("fit"); v != nil && !goja.IsUndefined(v) {
		opts.Fit = Fit(v.String())
	}
	if v := obj.Get("format"); v != nil && !goja.IsUndefined(v) {
		f := v.String()
		switch f {
		case "jpeg", "jpg":
			opts.Format = FormatJPEG
		case "png":
			opts.Format = FormatPNG
		}
	}
	if v := obj.Get("quality"); v != nil && !goja.IsUndefined(v) {
		opts.Quality = int(v.ToInteger())
	}

	return opts
}

// resultToJS converts a Result to a Goja JS object.
func resultToJS(vm *goja.Runtime, r *Result) goja.Value {
	obj := vm.NewObject()
	obj.Set("data", vm.NewArrayBuffer(r.Data))
	obj.Set("width", r.Width)
	obj.Set("height", r.Height)
	obj.Set("size", len(r.Data))
	return obj
}
