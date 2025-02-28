// Code generated by winrt-go-gen. DO NOT EDIT.

//go:build windows

//nolint:all
package foundation

import (
	"sync"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/internal/delegate"
	"github.com/saltosystems/winrt-go/internal/kernel32"
)

const GUIDDeferralCompletedHandler string = "ed32a372-f3c8-4faa-9cfb-470148da3888"
const SignatureDeferralCompletedHandler string = "delegate({ed32a372-f3c8-4faa-9cfb-470148da3888})"

type DeferralCompletedHandler struct {
	ole.IUnknown
	sync.Mutex
	refs uint64
	IID  ole.GUID
}

type DeferralCompletedHandlerVtbl struct {
	ole.IUnknownVtbl
	Invoke uintptr
}

type DeferralCompletedHandlerCallback func(instance *DeferralCompletedHandler)

var callbacksDeferralCompletedHandler = &deferralCompletedHandlerCallbacks{
	mu:        &sync.Mutex{},
	callbacks: make(map[unsafe.Pointer]DeferralCompletedHandlerCallback),
}

var releaseChannelsDeferralCompletedHandler = &deferralCompletedHandlerReleaseChannels{
	mu:    &sync.Mutex{},
	chans: make(map[unsafe.Pointer]chan struct{}),
}

func NewDeferralCompletedHandler(iid *ole.GUID, callback DeferralCompletedHandlerCallback) *DeferralCompletedHandler {
	// create type instance
	size := unsafe.Sizeof(*(*DeferralCompletedHandler)(nil))
	instPtr := kernel32.Malloc(size)
	inst := (*DeferralCompletedHandler)(instPtr)

	// get the callbacks for the VTable
	callbacks := delegate.RegisterCallbacks(instPtr, inst)

	// the VTable should also be allocated in the heap
	sizeVTable := unsafe.Sizeof(*(*DeferralCompletedHandlerVtbl)(nil))
	vTablePtr := kernel32.Malloc(sizeVTable)

	inst.RawVTable = (*interface{})(vTablePtr)

	vTable := (*DeferralCompletedHandlerVtbl)(vTablePtr)
	vTable.IUnknownVtbl = ole.IUnknownVtbl{
		QueryInterface: callbacks.QueryInterface,
		AddRef:         callbacks.AddRef,
		Release:        callbacks.Release,
	}
	vTable.Invoke = callbacks.Invoke

	// Initialize all properties: the malloc may contain garbage
	inst.IID = *iid // copy contents
	inst.Mutex = sync.Mutex{}
	inst.refs = 0

	callbacksDeferralCompletedHandler.add(unsafe.Pointer(inst), callback)

	// See the docs in the releaseChannelsDeferralCompletedHandler struct
	releaseChannelsDeferralCompletedHandler.acquire(unsafe.Pointer(inst))

	inst.addRef()
	return inst
}

func (r *DeferralCompletedHandler) GetIID() *ole.GUID {
	return &r.IID
}

// addRef increments the reference counter by one
func (r *DeferralCompletedHandler) addRef() uint64 {
	r.Lock()
	defer r.Unlock()
	r.refs++
	return r.refs
}

// removeRef decrements the reference counter by one. If it was already zero, it will just return zero.
func (r *DeferralCompletedHandler) removeRef() uint64 {
	r.Lock()
	defer r.Unlock()

	if r.refs > 0 {
		r.refs--
	}

	return r.refs
}

func (instance *DeferralCompletedHandler) Invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8 unsafe.Pointer) uintptr {

	// See the quote above.
	if callback, ok := callbacksDeferralCompletedHandler.get(instancePtr); ok {
		callback(instance)
	}
	return ole.S_OK
}

func (instance *DeferralCompletedHandler) AddRef() uint64 {
	return instance.addRef()
}

func (instance *DeferralCompletedHandler) Release() uint64 {
	rem := instance.removeRef()
	if rem == 0 {
		// We're done.
		instancePtr := unsafe.Pointer(instance)
		callbacksDeferralCompletedHandler.delete(instancePtr)

		// stop release channels used to avoid
		// https://github.com/golang/go/issues/55015
		releaseChannelsDeferralCompletedHandler.release(instancePtr)

		kernel32.Free(unsafe.Pointer(instance.RawVTable))
		kernel32.Free(instancePtr)
	}
	return rem
}

type deferralCompletedHandlerCallbacks struct {
	mu        *sync.Mutex
	callbacks map[unsafe.Pointer]DeferralCompletedHandlerCallback
}

func (m *deferralCompletedHandlerCallbacks) add(p unsafe.Pointer, v DeferralCompletedHandlerCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callbacks[p] = v
}

func (m *deferralCompletedHandlerCallbacks) get(p unsafe.Pointer) (DeferralCompletedHandlerCallback, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.callbacks[p]
	return v, ok
}

func (m *deferralCompletedHandlerCallbacks) delete(p unsafe.Pointer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.callbacks, p)
}

// typedEventHandlerReleaseChannels keeps a map with channels
// used to keep a goroutine alive during the lifecycle of this object.
// This is required to avoid causing a deadlock error.
// See this: https://github.com/golang/go/issues/55015
type deferralCompletedHandlerReleaseChannels struct {
	mu    *sync.Mutex
	chans map[unsafe.Pointer]chan struct{}
}

func (m *deferralCompletedHandlerReleaseChannels) acquire(p unsafe.Pointer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c := make(chan struct{})
	m.chans[p] = c

	go func() {
		// we need a timer to trick the go runtime into
		// thinking there's still something going on here
		// but we are only really interested in <-c
		t := time.NewTimer(time.Minute)
		for {
			select {
			case <-t.C:
				t.Reset(time.Minute)
			case <-c:
				t.Stop()
				return
			}
		}
	}()
}

func (m *deferralCompletedHandlerReleaseChannels) release(p unsafe.Pointer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.chans[p]; ok {
		close(c)
		delete(m.chans, p)
	}
}
