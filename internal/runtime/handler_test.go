package runtime

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
	"time"
)

func TestBuildRequest_MultipartFiles(t *testing.T) {
	// Create a multipart body with a file and a text field
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add a text field
	writer.WriteField("caption", "My photo")

	// Add a file
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="photo"; filename="test.png"`)
	partHeader.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatal(err)
	}
	fileData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic bytes
	part.Write(fileData)
	writer.Close()

	req, err := http.NewRequest("POST", "/api/upload", &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	result := buildRequest(req)

	// Check text field in body
	body, ok := result.Body.(map[string]string)
	if !ok {
		t.Fatalf("expected body to be map[string]string, got %T", result.Body)
	}
	if body["caption"] != "My photo" {
		t.Errorf("expected caption 'My photo', got %q", body["caption"])
	}

	// Check file
	if result.Files == nil {
		t.Fatal("expected files to be populated")
	}
	photo, ok := result.Files["photo"]
	if !ok {
		t.Fatal("expected 'photo' field in files")
	}
	if photo.Name != "test.png" {
		t.Errorf("expected filename 'test.png', got %q", photo.Name)
	}
	if photo.Type != "image/png" {
		t.Errorf("expected type 'image/png', got %q", photo.Type)
	}
	if photo.Size != len(fileData) {
		t.Errorf("expected size %d, got %d", len(fileData), photo.Size)
	}
	if !bytes.Equal(photo.Data, fileData) {
		t.Errorf("file data mismatch")
	}
}

func TestBuildRequest_MultipartNoFile(t *testing.T) {
	// Multipart with only text fields, no files
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "Alice")
	writer.WriteField("age", "30")
	writer.Close()

	req, err := http.NewRequest("POST", "/api/form", &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	result := buildRequest(req)

	body, ok := result.Body.(map[string]string)
	if !ok {
		t.Fatalf("expected body to be map[string]string, got %T", result.Body)
	}
	if body["name"] != "Alice" {
		t.Errorf("expected name 'Alice', got %q", body["name"])
	}
	if body["age"] != "30" {
		t.Errorf("expected age '30', got %q", body["age"])
	}

	// Files should be nil when no files uploaded
	if result.Files != nil {
		t.Errorf("expected no files, got %v", result.Files)
	}
}

func TestBuildRequest_JSONBody(t *testing.T) {
	// Ensure JSON parsing still works
	body := bytes.NewBufferString(`{"key":"value"}`)
	req, err := http.NewRequest("POST", "/api/test", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	result := buildRequest(req)

	bodyMap, ok := result.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected body to be map[string]interface{}, got %T", result.Body)
	}
	if bodyMap["key"] != "value" {
		t.Errorf("expected key 'value', got %v", bodyMap["key"])
	}
	if result.Files != nil {
		t.Error("expected no files for JSON request")
	}
}

func TestFileUpload_ArrayBufferInVM(t *testing.T) {
	// Test that files are injected as ArrayBuffer in the VM
	rt := NewRuntime(1, 2*time.Second)
	ctx := context.Background()

	fileData := []byte("hello world binary data")
	req := &Request{
		Method:  "POST",
		Path:    "/api/upload",
		Query:   map[string]string{},
		Headers: map[string]string{},
		Files: map[string]FileUpload{
			"doc": {
				Name: "readme.txt",
				Type: "text/plain",
				Size: len(fileData),
				Data: fileData,
			},
		},
	}

	// JS code that reads the file properties
	code := `
		var f = request.files.doc;
		({
			status: 200,
			body: {
				name: f.name,
				type: f.type,
				size: f.size,
				isArrayBuffer: f.data instanceof ArrayBuffer,
				dataLength: f.data.byteLength
			}
		})
	`

	result := rt.Execute(ctx, code, req)
	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}

	body, ok := result.Response.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected body to be map, got %T", result.Response.Body)
	}

	if body["name"] != "readme.txt" {
		t.Errorf("expected name 'readme.txt', got %v", body["name"])
	}
	if body["type"] != "text/plain" {
		t.Errorf("expected type 'text/plain', got %v", body["type"])
	}
	if size, ok := body["size"].(int64); !ok || size != int64(len(fileData)) {
		t.Errorf("expected size %d, got %v", len(fileData), body["size"])
	}
	if isAB, ok := body["isArrayBuffer"].(bool); !ok || !isAB {
		t.Errorf("expected data to be ArrayBuffer, got %v", body["isArrayBuffer"])
	}
	if dl, ok := body["dataLength"].(int64); !ok || dl != int64(len(fileData)) {
		t.Errorf("expected dataLength %d, got %v", len(fileData), body["dataLength"])
	}
}
