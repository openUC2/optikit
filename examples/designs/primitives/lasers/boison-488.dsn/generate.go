package example

//go:generate -command py uv run poe

//go:generate pwd

//go:generate ./generate-assm.sh generate-assm.directives
//go:generate py config _assembly:wiring.yml _newswitch-config.yml
