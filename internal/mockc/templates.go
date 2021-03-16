package mockc

import (
	"bytes"
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/dave/jennifer/jen"
)

func render(pkg *packages.Package, mocks []mockInfo, gogenerate string) ([]byte, error) {
	f := jen.NewFilePathName(pkg.PkgPath, pkg.Name)

	f.PackageComment("// Code generated by Mockc. DO NOT EDIT.")
	f.PackageComment("// repo: https://github.com/KimMachineGun/mockc\n")
	f.PackageComment(fmt.Sprintf("//go:generate %s", gogenerate))
	f.PackageComment("// +build !mockc\n")

	for _, mock := range mocks {
		f.Var().Id("_").Do(func(s *jen.Statement) {
			typeCode(s, mock.typ)
		}).Op("=").Op("&").Id(mock.name).Values()
		f.Type().Id(mock.name).StructFunc(func(g *jen.Group) {
			for _, method := range mock.methods {
				g.Commentf("method: %s", method.typ.Name())
				g.Id(method.fieldName).StructFunc(func(g *jen.Group) {
					g.Id("mu").Qual("sync", "Mutex")
					g.Comment("basics")
					g.Id("Called").Bool()
					g.Id("CallCount").Int()
					if len(method.params)+len(method.results) > 0 {
						g.Comment("call history")
						g.Id("History").Index().StructFunc(func(g *jen.Group) {
							if len(method.params) > 0 {
								g.Id("Params").StructFunc(func(g *jen.Group) {
									for i, param := range method.params {
										param := param
										g.Do(func(s *jen.Statement) {
											typeCode(s.Id(fmt.Sprintf("P%d", i)), param.typ.Type())
										})
									}
								})
							}
							if len(method.results) > 0 {
								g.Id("Results").StructFunc(func(g *jen.Group) {
									for i, result := range method.results {
										result := result
										g.Do(func(s *jen.Statement) {
											typeCode(s.Id(fmt.Sprintf("R%d", i)), result.typ.Type())
										})
									}
								})
							}
						})
					}
					if len(method.params) > 0 {
						g.Comment("params")
						g.Id("Params").StructFunc(func(g *jen.Group) {
							for i, param := range method.params {
								param := param
								g.Do(func(s *jen.Statement) {
									typeCode(s.Id(fmt.Sprintf("P%d", i)), param.typ.Type())
								})
							}
						})
					}
					if len(method.results) > 0 {
						g.Comment("results")
						g.Id("Results").StructFunc(func(g *jen.Group) {
							for i, result := range method.results {
								result := result
								g.Do(func(s *jen.Statement) {
									typeCode(s.Id(fmt.Sprintf("R%d", i)), result.typ.Type())
								})
							}
						})
					}
					g.Comment("if it is not nil, it'll be called in the middle of the method.")
					g.Id("Body").Do(func(s *jen.Statement) {
						typeCode(s, method.typ.Type())
					})
				})
			}
		})

		if mock.hasConstructor {
			f.Func().Id(fmt.Sprintf("New%s", mock.name)).Params(
				jen.Id("v").Op("...").Do(func(s *jen.Statement) {
					typeCode(s, mock.typ)
				}),
			).Op("*").Id(mock.name).Block(
				jen.Id("m").Op(":=").Op("&").Id(mock.name).Values(),
				jen.If(jen.Len(jen.Id("v")).Op(">").Lit(0)).BlockFunc(func(g *jen.Group) {
					for _, method := range mock.methods {
						g.Id("m").Dot(method.fieldName).Dot("Body").Op("=").Id("v").Index(jen.Lit(0)).Dot(method.typ.Name())
					}
				}),
				jen.Return(jen.Id("m")),
			).Line()
		}

		for _, method := range mock.methods {
			f.Func().Params(jen.Id("recv").Op("*").Id(mock.name)).Id(method.typ.Name()).ParamsFunc(func(g *jen.Group) {
				for i, param := range method.params {
					param := param
					g.Do(func(s *jen.Statement) {
						s.Id(fmt.Sprintf("p%d", i))
						if param.isVariadic {
							typeCode(s.Op("..."), param.typ.Type().(*types.Slice).Elem())
						} else {
							typeCode(s, param.typ.Type())
						}
					})
				}
			}).ParamsFunc(func(g *jen.Group) {
				for _, result := range method.results {
					result := result
					g.Do(func(s *jen.Statement) {
						typeCode(s, result.typ.Type())
					})
				}
			}).BlockFunc(func(g *jen.Group) {
				fieldName := jen.Id("recv").Dot(method.fieldName)

				g.Add(fieldName).Dot("mu").Dot("Lock").Call()
				g.Defer().Add(fieldName).Dot("mu").Dot("Unlock").Call()

				g.Comment("basics")
				g.Add(fieldName).Dot("Called").Op("=").True()
				g.Add(fieldName).Dot("CallCount").Op("++")

				if len(method.params) > 0 {
					g.Comment("params")
					for i := range method.params {
						g.Add(fieldName).Dot("Params").Dot(fmt.Sprintf("P%d", i)).Op("=").Id(fmt.Sprintf("p%d", i))
					}
				}

				g.Comment("body")
				g.If(jen.Add(fieldName).Dot("Body").Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
					g.Do(func(s *jen.Statement) {
						if len(method.results) > 0 {
							s.ListFunc(func(g *jen.Group) {
								for i := range method.results {
									g.Add(fieldName).Dot("Results").Dot(fmt.Sprintf("R%d", i))
								}
							}).Op("=")
						}
						s.Add(fieldName).Dot("Body").CallFunc(func(g *jen.Group) {
							for i, param := range method.params {
								g.Do(func(s *jen.Statement) {
									s.Id(fmt.Sprintf("p%d", i))
									if param.isVariadic {
										s.Op("...")
									}
								})
							}
						})
					})
				})

				if len(method.params)+len(method.results) > 0 {
					g.Comment("call history")
					g.Id("recv").Dot(method.fieldName).Dot("History").Op("=").Append(
						jen.Id("recv").Dot(method.fieldName).Dot("History"),
						jen.StructFunc(func(g *jen.Group) {
							if len(method.params) > 0 {
								g.Id("Params").StructFunc(func(g *jen.Group) {
									for i, param := range method.params {
										param := param
										g.Do(func(s *jen.Statement) {
											typeCode(s.Id(fmt.Sprintf("P%d", i)), param.typ.Type())
										})
									}
								})
							}
							if len(method.results) > 0 {
								g.Id("Results").StructFunc(func(g *jen.Group) {
									for i, result := range method.results {
										result := result
										g.Do(func(s *jen.Statement) {
											typeCode(s.Id(fmt.Sprintf("R%d", i)), result.typ.Type())
										})
									}
								})
							}
						}).Values(jen.DictFunc(func(d jen.Dict) {
							if len(method.params) > 0 {
								d[jen.Id("Params")] = jen.Add(fieldName).Dot("Params")
							}
							if len(method.results) > 0 {
								d[jen.Id("Results")] = jen.Add(fieldName).Dot("Results")
							}
						})),
					)
				}

				if len(method.results) > 0 {
					g.Comment("results")
					g.ReturnFunc(func(g *jen.Group) {
						for i := range method.results {
							g.Add(fieldName).Dot("Results").Dot(fmt.Sprintf("R%d", i))
						}
					})
				}
			}).Line()
		}
	}

	b := bytes.NewBuffer(nil)
	err := f.Render(b)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

type mockInfo struct {
	typ            *types.Interface
	name           string
	hasConstructor bool
	methods        []methodInfo
}

type methodInfo struct {
	typ       *types.Func
	fieldName string
	params    []paramInfo
	results   []resultInfo
}

type paramInfo struct {
	typ        *types.Var
	isVariadic bool
}

type resultInfo struct {
	typ *types.Var
}
