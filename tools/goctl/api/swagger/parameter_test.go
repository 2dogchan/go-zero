package swagger

import (
	"net/http"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	apiSpec "github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func TestIsPostJson(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		hasJson  bool
		expected bool
	}{
		{"POST with JSON", http.MethodPost, true, true},
		{"POST without JSON", http.MethodPost, false, false},
		{"GET with JSON", http.MethodGet, true, false},
		{"PUT with JSON", http.MethodPut, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := createTestStruct("TestStruct", tt.hasJson)
			_, result := isPostJson(testingContext(t), tt.method, testStruct)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParametersFromType(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		useDefinitions bool
		hasJson        bool
		expectedCount  int
		expectedBody   bool
	}{
		{"POST JSON with definitions", http.MethodPost, true, true, 1, true},
		{"POST JSON without definitions", http.MethodPost, false, true, 1, true},
		{"GET with form", http.MethodGet, false, false, 1, false},
		{"POST with form", http.MethodPost, false, false, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context{UseDefinitions: tt.useDefinitions}
			testStruct := createTestStruct("TestStruct", tt.hasJson)
			params := parametersFromType(ctx, tt.method, testStruct, apiSpec.AtDoc{})

			assert.Equal(t, tt.expectedCount, len(params))
			if tt.expectedBody {
				assert.Equal(t, paramsInBody, params[0].In)
			} else if len(params) > 0 {
				assert.NotEqual(t, paramsInBody, params[0].In)
			}
		})
	}
}

func TestParametersFromType_EdgeCases(t *testing.T) {
	ctx := testingContext(t)

	params := parametersFromType(ctx, http.MethodPost, nil, apiSpec.AtDoc{})
	assert.Empty(t, params)

	primitiveType := apiSpec.PrimitiveType{RawName: "string"}
	params = parametersFromType(ctx, http.MethodPost, primitiveType, apiSpec.AtDoc{})
	assert.Empty(t, params)
}

func TestParametersFromType_ArrayType(t *testing.T) {
	ctx := testingContext(t)

	// 测试 []int64 数组类型
	arrayType := apiSpec.ArrayType{
		RawName: "[]int64",
		Value: apiSpec.PrimitiveType{
			RawName: "int64",
		},
	}

	params := parametersFromType(ctx, http.MethodPost, arrayType, apiSpec.AtDoc{})

	// 应该返回一个参数
	assert.Equal(t, 1, len(params))

	// 参数应该是 body 类型
	param := params[0]
	assert.Equal(t, paramsInBody, param.In)
	assert.Equal(t, paramsInBody, param.Name)
	assert.True(t, param.Required)

	// Schema 应该是数组类型
	assert.NotNil(t, param.Schema)
	assert.Equal(t, spec.StringOrArray{swaggerTypeArray}, param.Schema.Type)

	// Items 应该存在
	assert.NotNil(t, param.Schema.Items)

	// Items 的 Schema 应该是 integer 类型
	itemSchema := param.Schema.Items.Schema
	assert.NotNil(t, itemSchema)
	assert.Equal(t, spec.StringOrArray{swaggerTypeInteger}, itemSchema.Type)
}

func TestParametersFromType_WithExample(t *testing.T) {
	ctx := testingContext(t)

	// 创建带有 example 标签的结构体
	testStruct := apiSpec.DefineStruct{
		RawName: "TestReq",
		Members: []apiSpec.Member{
			{
				Name: "Ids",
				Type: apiSpec.ArrayType{
					RawName: "[]int64",
					Value: apiSpec.PrimitiveType{
						RawName: "int64",
					},
				},
				Tag: `json:"ids" example:"1,2,3"`,
			},
			{
				Name: "Name",
				Type: apiSpec.PrimitiveType{
					RawName: "string",
				},
				Tag: `json:"name" example:"test"`,
			},
			{
				Name: "Count",
				Type: apiSpec.PrimitiveType{
					RawName: "int",
				},
				Tag: `json:"count" example:"10"`,
			},
		},
	}

	params := parametersFromType(ctx, http.MethodPost, testStruct, apiSpec.AtDoc{})

	// 应该返回一个 body 参数
	assert.Equal(t, 1, len(params))
	param := params[0]
	assert.Equal(t, paramsInBody, param.In)

	// Schema 应该包含 properties
	assert.NotNil(t, param.Schema)
	assert.NotNil(t, param.Schema.Properties)

	// 检查 Ids 字段的示例值
	idsSchema := param.Schema.Properties["ids"]
	assert.NotNil(t, idsSchema)
	assert.Equal(t, []any{int64(1), int64(2), int64(3)}, idsSchema.Example)

	// 检查 Name 字段的示例值
	nameSchema := param.Schema.Properties["name"]
	assert.NotNil(t, nameSchema)
	assert.Equal(t, "test", nameSchema.Example)

	// 检查 Count 字段的示例值
	countSchema := param.Schema.Properties["count"]
	assert.NotNil(t, countSchema)
	assert.Equal(t, int64(10), countSchema.Example)
}

func TestParametersFromType_WithAtDocExample(t *testing.T) {
	ctx := testingContext(t)

	// 测试数组类型
	arrayType := apiSpec.ArrayType{
		RawName: "[]int64",
		Value: apiSpec.PrimitiveType{
			RawName: "int64",
		},
	}

	// 创建带有 reqExample 的 AtDoc
	atDoc := apiSpec.AtDoc{
		Properties: map[string]string{
			propertyKeyReqExample: "[1,2,3,4,5]",
		},
	}

	params := parametersFromType(ctx, http.MethodPost, arrayType, atDoc)

	// 应该返回一个参数
	assert.Equal(t, 1, len(params))

	// 参数应该是 body 类型
	param := params[0]
	assert.Equal(t, paramsInBody, param.In)
	assert.True(t, param.Required)

	// Schema 应该是数组类型
	assert.NotNil(t, param.Schema)
	assert.Equal(t, spec.StringOrArray{swaggerTypeArray}, param.Schema.Type)

	// 检查示例值
	assert.NotNil(t, param.Schema.Example)
	assert.Equal(t, []any{int64(1), int64(2), int64(3), int64(4), int64(5)}, param.Schema.Example)
}

func TestJsonResponseFromType_WithAtDocExample(t *testing.T) {
	ctx := testingContext(t)

	// 测试数组类型
	arrayType := apiSpec.ArrayType{
		RawName: "[]int64",
		Value: apiSpec.PrimitiveType{
			RawName: "int64",
		},
	}

	// 创建带有 respExample 的 AtDoc
	atDoc := apiSpec.AtDoc{
		Properties: map[string]string{
			propertyKeyRespExample: "[10,20,30]",
		},
	}

	responses := jsonResponseFromType(ctx, atDoc, arrayType)

	// 应该返回响应
	assert.NotNil(t, responses)
	assert.NotNil(t, responses.StatusCodeResponses)

	// 检查 200 响应
	response, exists := responses.StatusCodeResponses[200]
	assert.True(t, exists)
	assert.NotNil(t, response.Schema)

	// 检查示例值
	assert.NotNil(t, response.Schema.Example)
	assert.Equal(t, []any{int64(10), int64(20), int64(30)}, response.Schema.Example)
}

func TestParseArrayExample(t *testing.T) {
	// 测试数组示例解析
	elementType := apiSpec.PrimitiveType{RawName: "int64"}

	result := parseArrayExample("1,2,3", elementType)
	t.Logf("Parse result: %+v (type: %T)", result, result)

	// 验证结果
	assert.NotNil(t, result)
	if arr, ok := result.([]any); ok {
		assert.Equal(t, 3, len(arr))
		assert.Equal(t, int64(1), arr[0])
		assert.Equal(t, int64(2), arr[1])
		assert.Equal(t, int64(3), arr[2])
	} else {
		t.Errorf("Expected []any, got %T", result)
	}
}

func createTestStruct(name string, hasJson bool) apiSpec.DefineStruct {
	tag := `form:"username"`
	if hasJson {
		tag = `json:"username"`
	}

	return apiSpec.DefineStruct{
		RawName: name,
		Members: []apiSpec.Member{
			{
				Name: "Username",
				Type: apiSpec.PrimitiveType{RawName: "string"},
				Tag:  tag,
			},
		},
	}
}
