// Package oodle decompresses Oodle (Kraken/Mermaid/Leviathan/…) data in pure Go
// with no CGO: it runs an embedded WebAssembly build of the open-source ooz
// decompressor on the pure-Go wazero runtime. This keeps the editor a single
// static, cross-compilable binary while reading the game's Oodle-compressed pak
// assets. See build/README.md for how ooz.wasm is produced.
package oodle

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed ooz.wasm
var wasmBin []byte

// safeSpace mirrors ooz's OOZ_SAFE_SPACE: the decoder may scribble past the end
// of the output buffer, so we over-allocate the destination.
const safeSpace = 64

// Decoder holds one instantiated ooz module. It is safe for concurrent use.
type Decoder struct {
	ctx     context.Context
	runtime wazero.Runtime
	mu      sync.Mutex
	mod     api.Module
	malloc  api.Function
	free    api.Function
	kraken  api.Function
	mem     api.Memory
}

// New instantiates the embedded ooz decompressor.
func New() (*Decoder, error) {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)
	// Emscripten's ALLOW_MEMORY_GROWTH notifies the host on grow; a no-op suffices.
	if _, err := rt.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(context.Context, uint32) {}).
		Export("emscripten_notify_memory_growth").
		Instantiate(ctx); err != nil {
		rt.Close(ctx)
		return nil, err
	}
	mod, err := rt.InstantiateWithConfig(ctx, wasmBin, wazero.NewModuleConfig())
	if err != nil {
		rt.Close(ctx)
		return nil, err
	}
	if init := mod.ExportedFunction("_initialize"); init != nil {
		if _, err := init.Call(ctx); err != nil {
			rt.Close(ctx)
			return nil, err
		}
	}
	d := &Decoder{
		ctx: ctx, runtime: rt, mod: mod,
		malloc: mod.ExportedFunction("malloc"),
		free:   mod.ExportedFunction("free"),
		kraken: mod.ExportedFunction("Kraken_Decompress"),
		mem:    mod.Memory(),
	}
	if d.malloc == nil || d.free == nil || d.kraken == nil || d.mem == nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("oodle: ooz.wasm missing expected exports")
	}
	return d, nil
}

// Close releases the runtime.
func (d *Decoder) Close() error { return d.runtime.Close(d.ctx) }

// Decompress decompresses one Oodle block to exactly rawSize bytes.
func (d *Decoder) Decompress(comp []byte, rawSize int) ([]byte, error) {
	if rawSize <= 0 {
		return nil, fmt.Errorf("oodle: bad raw size %d", rawSize)
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	srcPtr, err := d.alloc(len(comp))
	if err != nil {
		return nil, err
	}
	defer d.free.Call(d.ctx, srcPtr)
	dstPtr, err := d.alloc(rawSize + safeSpace)
	if err != nil {
		return nil, err
	}
	defer d.free.Call(d.ctx, dstPtr)

	if !d.mem.Write(uint32(srcPtr), comp) {
		return nil, fmt.Errorf("oodle: write out of range")
	}
	res, err := d.kraken.Call(d.ctx, srcPtr, uint64(len(comp)), dstPtr, uint64(rawSize))
	if err != nil {
		return nil, fmt.Errorf("oodle: decode: %w", err)
	}
	if n := int(int32(res[0])); n != rawSize {
		return nil, fmt.Errorf("oodle: decoded %d bytes, want %d", n, rawSize)
	}
	out, ok := d.mem.Read(uint32(dstPtr), uint32(rawSize))
	if !ok {
		return nil, fmt.Errorf("oodle: read out of range")
	}
	return append([]byte(nil), out...), nil
}

func (d *Decoder) alloc(n int) (uint64, error) {
	r, err := d.malloc.Call(d.ctx, uint64(n))
	if err != nil {
		return 0, err
	}
	if r[0] == 0 {
		return 0, fmt.Errorf("oodle: malloc(%d) failed", n)
	}
	return r[0], nil
}
