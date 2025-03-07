package dcom

// https://learn.microsoft.com/en-us/windows/win32/api/wtypesbase/ne-wtypesbase-clsctx
type ClassContext uint32

const (
	// The code that creates and manages objects of this class is a DLL that runs
	// in the same process as the caller of the function specifying the class context.
	ClassContextInprocServer ClassContext = 0x1
	// The code that manages objects of this class is an in-process handler.
	// This is a DLL that runs in the client process and implements client-side
	// structures of this class when instances of the class are accessed remotely.
	ClassContextInprocHandler ClassContext = 0x2
	// The EXE code that creates and manages objects of this class runs on
	// same machine but is loaded in a separate process space.
	ClassContextLocalServer ClassContext = 0x4
	// A remote context. The LocalServer32 or LocalService code that creates
	// and manages objects of this class is run on a different computer.
	ClassContextRemoteServer ClassContext = 0x10000
)
