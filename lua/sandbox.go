package lua

import (
	"github.com/aarzilli/golua/lua"
	"github.com/stevedonovan/luar"
)

type MetaMap map[string]func(*lua.State) int

func CreateSandbox(state *lua.State, meta MetaMap) {

	state.CreateTable(0, len(globals))

	for _, g := range globals {
		state.GetGlobal(g)
		state.SetField(-2, g)
	}

	if meta != nil {
		state.CreateTable(0, len(meta))
		for k, m := range meta {
			state.SetMetaMethod(k, m)
		}
		state.SetMetaTable(-2)
	}

}

func CreateTable(state *lua.State, m luar.Map, meta MetaMap) {
	state.CreateTable(0, len(m))
	for k, g := range m {
		luar.GoToLua(state, g)
		state.SetField(-2, k)
	}

	if meta != nil {
		state.CreateTable(0, len(meta))
		for k, m := range meta {
			state.SetMetaMethod(k, m)
		}
		state.SetMetaTable(-2)
	}
}
