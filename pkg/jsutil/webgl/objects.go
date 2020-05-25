package webgl

// ObjectData vertices
type ObjectData struct {
	// vec4 in shader?
	Positions []float32
	Normals   []float32
	TexCoords []float32
	Indices   []uint16
}
