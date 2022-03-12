/**
@copyright: fantasysky 2016
@brief: runtime utils
@author: fanky
@version: 1.0
@date: 2019-07-01
**/

package fsruntime

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

type s_Functab struct {
	entry   uintptr
	funcoff uintptr
}

type s_Textsect struct {
	vaddr    uintptr // prelinked section vaddr
	length   uintptr // section length
	baseaddr uintptr // relocated section address
}

type s_Itab struct {
	inter *interfacetype
	_type *_type
	hash  uint32 // copy of _type.hash. Used for type switches.
	_     [4]byte
	fun   [1]uintptr // variable sized. fun[0]==0 means _type does not implement inter.
}

// A ptabEntry is generated by the compiler for each exported function
// and global variable in the main package of a plugin. It is used to
// initialize the plugin module's symbol map.
type s_PtabEntry struct {
	name nameOff
	typ  typeOff
}

// just like runtime/symtab.go::modulehash
type s_Modulehash struct {
	modulename   string
	linktimehash string
	runtimehash  *string
}

// just like runtime/stack.go::bitvector
type s_Bitvector struct {
	n        int32 // # of bits
	bytedata *uint8
}

type s_Moduledata struct {
	pclntable    []byte
	ftab         []s_Functab
	filetab      []uint32
	findfunctab  uintptr
	minpc, maxpc uintptr

	text, etext           uintptr
	noptrdata, enoptrdata uintptr
	data, edata           uintptr
	bss, ebss             uintptr
	noptrbss, enoptrbss   uintptr
	end, gcdata, gcbss    uintptr
	types, etypes         uintptr

	textsectmap []s_Textsect
	typelinks   []int32 // offsets from types
	itablinks   []*s_Itab

	ptab []s_PtabEntry

	pluginpath string
	pkghashes  []s_Modulehash

	modulename   string
	modulehashes []s_Modulehash

	hasmain uint8 // 1 if module contains the main function, 0 otherwise

	gcdatamask, gcbssmask s_Bitvector

	typemap map[typeOff]*_type // offset to *_rtype in previous module

	bad bool // module failed to load and should be ignored

	next *s_Moduledata
}

//go:linkname g_moduleData runtime/firstmoduledata
var g_moduleData s_Moduledata

func FindFuncWithName(name string) (uintptr, error) {
	//m := (*s_Moduledata)(unsafe.Pointer(runtime.FirstModuleData()))
	for moduleData := &g_moduleData; moduleData != nil; moduleData = moduleData.next {
		for _, ftab := range moduleData.ftab {
			f := (*runtime.Func)(unsafe.Pointer(&moduleData.pclntable[ftab.funcoff]))
			if f.Name() == name {
				return f.Entry(), nil
			}
		}
	}
	return 0, fmt.Errorf("Invalid function name: %s", name)
}

// 获取函数的导入名称，返回格式如："fsruntime.GetFuncName"
// 获取失败（传入非函数），则返回空字符串
func GetFuncName(fun interface{}) string {
	vfun := reflect.ValueOf(fun)
	if vfun.Kind() != reflect.Func {
		return ""
	}
	return runtime.FuncForPC(vfun.Pointer()).Name()
}