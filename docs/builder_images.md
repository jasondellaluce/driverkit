# Builder Images

Driverkit supports multiple builder images.  
A builder image is the docker image used to build the drivers.

## Adding a builder image

Adding a builder image is just a matter of adding a new dockerfile under the [build](../build) folder,  
with a name matching: `builder_$osname.Dockerfile` (like: `builder_stretch.Dockerfile`).  

The makefile will be then automatically able to collect the new docker images and pushing it as part of the CI.  

Moreover, the new image $osname must also be added to the static map of images, kept in [builders.go source file](../pkg/driverbuilder/builder/builders.go):  
```go
var images = map[string]Image{
	"buster": {
		GCCVersion: []float64{4.8, 5, 6, 8},
	},
	"bullseye": {
		GCCVersion: []float64{9, 10},
	},
	"bookworm": {
		GCCVersion: []float64{11, 12},
	},
	"stretch": {
        GCCVersion: []float64{6},
    },
}
```

Then, the new image's shipped gcc is now available to various builder using the `defaultGCC` method's algorithm,  
or chosen by each builder by implementing the `builder.GCCVersionRequestor` interface.  

Finally, the [builder](builder.md) doc file should be updated with the new available GCC (section 3.).