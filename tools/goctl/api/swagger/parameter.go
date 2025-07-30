package swagger

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	apiSpec "github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func isPostJson(ctx Context, method string, tp apiSpec.Type) (string, bool) {
	if !strings.EqualFold(method, http.MethodPost) {
		return "", false
	}
	structType, ok := tp.(apiSpec.DefineStruct)
	if !ok {
		return "", false
	}
	var isPostJson bool
	rangeMemberAndDo(ctx, structType, func(tag *apiSpec.Tags, required bool, member apiSpec.Member) {
		jsonTag, _ := tag.Get(tagJson)
		if !isPostJson {
			isPostJson = jsonTag != nil
		}
	})
	return structType.RawName, isPostJson
}

func parametersFromType(ctx Context, method string, tp apiSpec.Type, atDoc apiSpec.AtDoc) []spec.Parameter {
	if tp == nil {
		return []spec.Parameter{}
	}

	// 处理数组类型，如 []int64
	if arrayType, ok := tp.(apiSpec.ArrayType); ok {
		// 对于数组类型，创建 body 参数
		param := spec.Parameter{
			ParamProps: spec.ParamProps{
				In:       paramsInBody,
				Name:     paramsInBody,
				Required: true,
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type:  []string{swaggerTypeArray},
						Items: itemsFromGoType(ctx, arrayType),
					},
				},
			},
		}

		// 检查是否有 reqExample
		if reqExample := getStringFromKVOrDefault(atDoc.Properties, propertyKeyReqExample, ""); reqExample != "" {
			exampleValue := parseExampleValue(reqExample, arrayType)
			if exampleValue != nil {
				param.Schema.Example = exampleValue
			}
		}

		return []spec.Parameter{param}
	}

	structType, ok := tp.(apiSpec.DefineStruct)
	if !ok {
		return []spec.Parameter{}
	}

	var (
		resp           []spec.Parameter
		properties     = map[string]spec.Schema{}
		requiredFields []string
	)
	rangeMemberAndDo(ctx, structType, func(tag *apiSpec.Tags, required bool, member apiSpec.Member) {
		headerTag, _ := tag.Get(tagHeader)
		hasHeader := headerTag != nil

		pathParameterTag, _ := tag.Get(tagPath)
		hasPathParameter := pathParameterTag != nil

		formTag, _ := tag.Get(tagForm)
		hasForm := formTag != nil

		jsonTag, _ := tag.Get(tagJson)
		hasJson := jsonTag != nil

		exampleTag, _ := tag.Get("example")
		hasExample := exampleTag != nil

		if hasHeader {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(headerTag.Options)
			resp = append(resp, spec.Parameter{
				CommonValidations: spec.CommonValidations{
					Maximum:          maximum,
					ExclusiveMaximum: exclusiveMaximum,
					Minimum:          minimum,
					ExclusiveMinimum: exclusiveMinimum,
					Enum:             enumsValueFromOptions(headerTag.Options),
				},
				SimpleSchema: spec.SimpleSchema{
					Type:    sampleTypeFromGoType(ctx, member.Type),
					Default: defValueFromOptions(ctx, headerTag.Options, member.Type),
					Items:   sampleItemsFromGoType(ctx, member.Type),
				},
				ParamProps: spec.ParamProps{
					In:          paramsInHeader,
					Name:        headerTag.Name,
					Description: formatComment(member.Comment),
					Required:    required,
				},
			})
		}
		if hasPathParameter {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(pathParameterTag.Options)
			resp = append(resp, spec.Parameter{
				CommonValidations: spec.CommonValidations{
					Maximum:          maximum,
					ExclusiveMaximum: exclusiveMaximum,
					Minimum:          minimum,
					ExclusiveMinimum: exclusiveMinimum,
					Enum:             enumsValueFromOptions(pathParameterTag.Options),
				},
				SimpleSchema: spec.SimpleSchema{
					Type:    sampleTypeFromGoType(ctx, member.Type),
					Default: defValueFromOptions(ctx, pathParameterTag.Options, member.Type),
					Items:   sampleItemsFromGoType(ctx, member.Type),
				},
				ParamProps: spec.ParamProps{
					In:          paramsInPath,
					Name:        pathParameterTag.Name,
					Description: formatComment(member.Comment),
					Required:    required,
				},
			})
		}
		if hasForm {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(formTag.Options)
			if strings.EqualFold(method, http.MethodGet) {
				resp = append(resp, spec.Parameter{
					CommonValidations: spec.CommonValidations{
						Maximum:          maximum,
						ExclusiveMaximum: exclusiveMaximum,
						Minimum:          minimum,
						ExclusiveMinimum: exclusiveMinimum,
						Enum:             enumsValueFromOptions(formTag.Options),
					},
					SimpleSchema: spec.SimpleSchema{
						Type:    sampleTypeFromGoType(ctx, member.Type),
						Default: defValueFromOptions(ctx, formTag.Options, member.Type),
						Items:   sampleItemsFromGoType(ctx, member.Type),
					},
					ParamProps: spec.ParamProps{
						In:              paramsInQuery,
						Name:            formTag.Name,
						Description:     formatComment(member.Comment),
						Required:        required,
						AllowEmptyValue: !required,
					},
				})
			} else {
				resp = append(resp, spec.Parameter{
					CommonValidations: spec.CommonValidations{
						Maximum:          maximum,
						ExclusiveMaximum: exclusiveMaximum,
						Minimum:          minimum,
						ExclusiveMinimum: exclusiveMinimum,
						Enum:             enumsValueFromOptions(formTag.Options),
					},
					SimpleSchema: spec.SimpleSchema{
						Type:    sampleTypeFromGoType(ctx, member.Type),
						Default: defValueFromOptions(ctx, formTag.Options, member.Type),
						Items:   sampleItemsFromGoType(ctx, member.Type),
					},
					ParamProps: spec.ParamProps{
						In:              paramsInForm,
						Name:            formTag.Name,
						Description:     formatComment(member.Comment),
						Required:        required,
						AllowEmptyValue: !required,
					},
				})
			}

		}
		if hasJson {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(jsonTag.Options)
			if required {
				requiredFields = append(requiredFields, jsonTag.Name)
			}
			var schema = spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					Example: exampleValueFromOptions(ctx, jsonTag.Options, member.Type),
				},
				SchemaProps: spec.SchemaProps{
					Description:          formatComment(member.Comment),
					Type:                 typeFromGoType(ctx, member.Type),
					Default:              defValueFromOptions(ctx, jsonTag.Options, member.Type),
					Maximum:              maximum,
					ExclusiveMaximum:     exclusiveMaximum,
					Minimum:              minimum,
					ExclusiveMinimum:     exclusiveMinimum,
					Enum:                 enumsValueFromOptions(jsonTag.Options),
					AdditionalProperties: mapFromGoType(ctx, member.Type),
				},
			}

			if hasExample {
				// 解析 example 标签的值
				exampleValue := parseExampleValueFromTag(exampleTag, member.Type)
				if exampleValue != nil {
					schema.Example = exampleValue
				}
			}

			switch sampleTypeFromGoType(ctx, member.Type) {
			case swaggerTypeArray:
				schema.Items = itemsFromGoType(ctx, member.Type)
			case swaggerTypeObject:
				p, r := propertiesFromType(ctx, member.Type)
				schema.Properties = p
				schema.Required = r
			}
			properties[jsonTag.Name] = schema
		}
	})
	if len(properties) > 0 {
		if ctx.UseDefinitions {
			structName, ok := isPostJson(ctx, method, tp)
			if ok {
				param := spec.Parameter{
					ParamProps: spec.ParamProps{
						In:       paramsInBody,
						Name:     paramsInBody,
						Required: true,
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: spec.MustCreateRef(getRefName(structName)),
							},
						},
					},
				}

				// 检查是否有 reqExample
				if reqExample := getStringFromKVOrDefault(atDoc.Properties, propertyKeyReqExample, ""); reqExample != "" {
					exampleValue := parseExampleValue(reqExample, tp)
					if exampleValue != nil {
						param.Schema.Example = exampleValue
					}
				}

				resp = append(resp, param)
			}
		} else {
			param := spec.Parameter{
				ParamProps: spec.ParamProps{
					In:       paramsInBody,
					Name:     paramsInBody,
					Required: true,
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type:       typeFromGoType(ctx, structType),
							Properties: properties,
							Required:   requiredFields,
						},
					},
				},
			}

			// 检查是否有 reqExample
			if reqExample := getStringFromKVOrDefault(atDoc.Properties, propertyKeyReqExample, ""); reqExample != "" {
				exampleValue := parseExampleValue(reqExample, tp)
				if exampleValue != nil {
					param.Schema.Example = exampleValue
				}
			}

			resp = append(resp, param)
		}
	}
	return resp
}

// parseExampleValue 解析 example 标签的值，根据字段类型生成合适的示例值
func parseExampleValue(exampleStr string, memberType apiSpec.Type) any {
	if exampleStr == "" {
		return nil
	}

	switch val := memberType.(type) {
	case apiSpec.PrimitiveType:
		return parsePrimitiveExample(exampleStr, val.RawName)
	case apiSpec.ArrayType:
		return parseArrayExample(exampleStr, val.Value)
	case apiSpec.MapType:
		// 对于 map 类型，返回字符串形式的示例
		return exampleStr
	case apiSpec.DefineStruct, apiSpec.NestedStruct:
		// 对于结构体类型，返回字符串形式的示例
		return exampleStr
	case apiSpec.PointerType:
		return parseExampleValue(exampleStr, val.Type)
	default:
		return exampleStr
	}
}

// parseExampleValueFromTag 从标签中解析示例值，处理 structtag 解析的结果
func parseExampleValueFromTag(exampleTag *apiSpec.Tag, memberType apiSpec.Type) any {
	if exampleTag == nil {
		return nil
	}

	// 组合 Name 和 Options 来重建完整的示例字符串
	var exampleStr string
	if exampleTag.Name != "" {
		exampleStr = exampleTag.Name
	}

	// 添加 Options 中的值
	for _, option := range exampleTag.Options {
		if exampleStr != "" {
			exampleStr += "," + option
		} else {
			exampleStr = option
		}
	}

	return parseExampleValue(exampleStr, memberType)
}

// parsePrimitiveExample 解析基本类型的示例值
func parsePrimitiveExample(exampleStr, typeName string) any {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64":
		if val, err := strconv.ParseInt(exampleStr, 10, 64); err == nil {
			return val
		}
	case "uint", "uint8", "uint16", "uint32", "uint64":
		if val, err := strconv.ParseUint(exampleStr, 10, 64); err == nil {
			return val
		}
	case "float32", "float64":
		if val, err := strconv.ParseFloat(exampleStr, 64); err == nil {
			return val
		}
	case "bool":
		if val, err := strconv.ParseBool(exampleStr); err == nil {
			return val
		}
	case "string":
		return exampleStr
	default:
		return exampleStr
	}
	return exampleStr
}

// parseArrayExample 解析数组类型的示例值
func parseArrayExample(exampleStr string, elementType apiSpec.Type) any {
	// 检查是否是 JSON 格式的数组字符串
	if strings.HasPrefix(exampleStr, "[") && strings.HasSuffix(exampleStr, "]") {
		// 移除方括号，然后按逗号分割
		content := strings.TrimSpace(exampleStr[1 : len(exampleStr)-1])
		if content == "" {
			return []any{}
		}

		parts := strings.Split(content, ",")
		var result []any
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			switch val := elementType.(type) {
			case apiSpec.PrimitiveType:
				parsed := parsePrimitiveExample(part, val.RawName)
				result = append(result, parsed)
			case apiSpec.ArrayType:
				// 对于嵌套数组，递归解析
				parsed := parseArrayExample(part, val.Value)
				result = append(result, parsed)
			default:
				result = append(result, part)
			}
		}
		return result
	}

	// 原有的逗号分割逻辑
	parts := strings.Split(exampleStr, ",")
	if len(parts) == 0 {
		return []any{}
	}

	var result []any
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		switch val := elementType.(type) {
		case apiSpec.PrimitiveType:
			parsed := parsePrimitiveExample(part, val.RawName)
			result = append(result, parsed)
		case apiSpec.ArrayType:
			// 对于嵌套数组，递归解析
			parsed := parseArrayExample(part, val.Value)
			result = append(result, parsed)
		default:
			result = append(result, part)
		}
	}

	return result
}
