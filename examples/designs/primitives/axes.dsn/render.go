package example

//go:generate -command geom go run ../../../../main.go dev dsn geom

//go:generate geom --variant=ZposXpos report-prim --format=yaml _primitives:ZposXpos.yml
//go:generate geom --variant=ZposXpos report-prim --format=json _primitives:ZposXpos.json
//go:generate geom --variant=ZposXpos render-obj _objects:ZposXpos.step
//go:generate geom --variant=ZposYpos report-prim --format=yaml _primitives:ZposYpos.yml
//go:generate geom --variant=ZposYpos report-prim --format=json _primitives:ZposYpos.json
//go:generate geom --variant=ZposYpos render-obj _objects:ZposYpos.step
//go:generate geom --variant=ZposXneg report-prim --format=yaml _primitives:ZposXneg.yml
//go:generate geom --variant=ZposXneg report-prim --format=json _primitives:ZposXneg.json
//go:generate geom --variant=ZposXneg render-obj _objects:ZposXneg.step
//go:generate geom --variant=ZposYneg report-prim --format=yaml _primitives:ZposYneg.yml
//go:generate geom --variant=ZposYneg report-prim --format=json _primitives:ZposYneg.json
//go:generate geom --variant=ZposYneg render-obj _objects:ZposYneg.step
//go:generate geom --variant=ZnegXpos report-prim --format=yaml _primitives:ZnegXpos.yml
//go:generate geom --variant=ZnegXpos report-prim --format=json _primitives:ZnegXpos.json
//go:generate geom --variant=ZnegXpos render-obj _objects:ZnegXpos.step
//go:generate geom --variant=ZnegYpos report-prim --format=yaml _primitives:ZnegYpos.yml
//go:generate geom --variant=ZnegYpos report-prim --format=json _primitives:ZnegYpos.json
//go:generate geom --variant=ZnegYpos render-obj _objects:ZnegYpos.step
//go:generate geom --variant=ZnegXneg report-prim --format=yaml _primitives:ZnegXneg.yml
//go:generate geom --variant=ZnegXneg report-prim --format=json _primitives:ZnegXneg.json
//go:generate geom --variant=ZnegXneg render-obj _objects:ZnegXneg.step
//go:generate geom --variant=ZnegYneg report-prim --format=yaml _primitives:ZnegYneg.yml
//go:generate geom --variant=ZnegYneg report-prim --format=json _primitives:ZnegYneg.json
//go:generate geom --variant=ZnegYneg render-obj _objects:ZnegYneg.step
//go:generate geom --variant=YposXpos report-prim --format=yaml _primitives:YposXpos.yml
//go:generate geom --variant=YposXpos report-prim --format=json _primitives:YposXpos.json
//go:generate geom --variant=YposXpos render-obj _objects:YposXpos.step
//go:generate geom --variant=YposZneg report-prim --format=yaml _primitives:YposZneg.yml
//go:generate geom --variant=YposZneg report-prim --format=json _primitives:YposZneg.json
//go:generate geom --variant=YposZneg render-obj _objects:YposZneg.step
//go:generate geom --variant=YposXneg report-prim --format=yaml _primitives:YposXneg.yml
//go:generate geom --variant=YposXneg report-prim --format=json _primitives:YposXneg.json
//go:generate geom --variant=YposXneg render-obj _objects:YposXneg.step
//go:generate geom --variant=YposZpos report-prim --format=yaml _primitives:YposZpos.yml
//go:generate geom --variant=YposZpos report-prim --format=json _primitives:YposZpos.json
//go:generate geom --variant=YposZpos render-obj _objects:YposZpos.step
//go:generate geom --variant=YnegXpos report-prim --format=yaml _primitives:YnegXpos.yml
//go:generate geom --variant=YnegXpos report-prim --format=json _primitives:YnegXpos.json
//go:generate geom --variant=YnegXpos render-obj _objects:YnegXpos.step
//go:generate geom --variant=YnegZneg report-prim --format=yaml _primitives:YnegZneg.yml
//go:generate geom --variant=YnegZneg report-prim --format=json _primitives:YnegZneg.json
//go:generate geom --variant=YnegZneg render-obj _objects:YnegZneg.step
//go:generate geom --variant=YnegXneg report-prim --format=yaml _primitives:YnegXneg.yml
//go:generate geom --variant=YnegXneg report-prim --format=json _primitives:YnegXneg.json
//go:generate geom --variant=YnegXneg render-obj _objects:YnegXneg.step
//go:generate geom --variant=YnegZpos report-prim --format=yaml _primitives:YnegZpos.yml
//go:generate geom --variant=YnegZpos report-prim --format=json _primitives:YnegZpos.json
//go:generate geom --variant=YnegZpos render-obj _objects:YnegZpos.step
//go:generate geom --variant=XposZneg report-prim --format=yaml _primitives:XposZneg.yml
//go:generate geom --variant=XposZneg report-prim --format=json _primitives:XposZneg.json
//go:generate geom --variant=XposZneg render-obj _objects:XposZneg.step
//go:generate geom --variant=XposYpos report-prim --format=yaml _primitives:XposYpos.yml
//go:generate geom --variant=XposYpos report-prim --format=json _primitives:XposYpos.json
//go:generate geom --variant=XposYpos render-obj _objects:XposYpos.step
//go:generate geom --variant=XposZpos report-prim --format=yaml _primitives:XposZpos.yml
//go:generate geom --variant=XposZpos report-prim --format=json _primitives:XposZpos.json
//go:generate geom --variant=XposZpos render-obj _objects:XposZpos.step
//go:generate geom --variant=XposYneg report-prim --format=yaml _primitives:XposYneg.yml
//go:generate geom --variant=XposYneg report-prim --format=json _primitives:XposYneg.json
//go:generate geom --variant=XposYneg render-obj _objects:XposYneg.step
//go:generate geom --variant=XnegZneg report-prim --format=yaml _primitives:XnegZneg.yml
//go:generate geom --variant=XnegZneg report-prim --format=json _primitives:XnegZneg.json
//go:generate geom --variant=XnegZneg render-obj _objects:XnegZneg.step
//go:generate geom --variant=XnegYpos report-prim --format=yaml _primitives:XnegYpos.yml
//go:generate geom --variant=XnegYpos report-prim --format=json _primitives:XnegYpos.json
//go:generate geom --variant=XnegYpos render-obj _objects:XnegYpos.step
//go:generate geom --variant=XnegZpos report-prim --format=yaml _primitives:XnegZpos.yml
//go:generate geom --variant=XnegZpos report-prim --format=json _primitives:XnegZpos.json
//go:generate geom --variant=XnegZpos render-obj _objects:XnegZpos.step
//go:generate geom --variant=XnegYneg report-prim --format=yaml _primitives:XnegYneg.yml
//go:generate geom --variant=XnegYneg report-prim --format=json _primitives:XnegYneg.json
//go:generate geom --variant=XnegYneg render-obj _objects:XnegYneg.step
