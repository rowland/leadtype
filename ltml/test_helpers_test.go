package ltml

import "testing/fstest"

func testingMapFS(name, data string) fstest.MapFS {
	return fstest.MapFS{
		name: &fstest.MapFile{Data: []byte(data)},
	}
}
