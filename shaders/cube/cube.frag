#version 300 es

precision mediump float;

in vec3 vColor;

// to follow the OpenGL spec.
out vec4 FragColor;

void main(void) {
	FragColor = vec4(vColor, 1.);
}
