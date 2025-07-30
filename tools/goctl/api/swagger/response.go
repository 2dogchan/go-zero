package swagger

import (
	"net/http"

	"github.com/go-openapi/spec"
	apiSpec "github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func jsonResponseFromType(ctx Context, atDoc apiSpec.AtDoc, tp apiSpec.Type) *spec.Responses {
	if tp == nil {
		return &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: map[int]spec.Response{
					http.StatusOK: {
						ResponseProps: spec.ResponseProps{
							Description: "",
							Schema:      &spec.Schema{},
						},
					},
				},
			},
		}
	}
	props := spec.SchemaProps{
		AdditionalProperties: mapFromGoType(ctx, tp),
		Items:                itemsFromGoType(ctx, tp),
	}

	// 检查是否有 respExample
	var example any
	if respExample := getStringFromKVOrDefault(atDoc.Properties, propertyKeyRespExample, ""); respExample != "" {
		example = parseExampleValue(respExample, tp)
	}

	if ctx.UseDefinitions {
		structName, ok := containsStruct(tp)
		if ok {
			props.Ref = spec.MustCreateRef(getRefName(structName))
			schema := &spec.Schema{
				SchemaProps: wrapCodeMsgProps(ctx, props, atDoc),
			}
			if example != nil {
				schema.Example = example
			}
			return &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusOK: {
							ResponseProps: spec.ResponseProps{
								Schema: schema,
							},
						},
					},
				},
			}
		}
	}

	p, _ := propertiesFromType(ctx, tp)
	props.Type = typeFromGoType(ctx, tp)
	props.Properties = p
	schema := &spec.Schema{
		SchemaProps: wrapCodeMsgProps(ctx, props, atDoc),
	}
	if example != nil {
		schema.Example = example
	}
	return &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				http.StatusOK: {
					ResponseProps: spec.ResponseProps{
						Schema: schema,
					},
				},
			},
		},
	}
}
