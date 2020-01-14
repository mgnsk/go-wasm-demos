#version 300 es

precision mediump float;

uniform vec3 vColor;

out vec4 fragColor;
void main(void) {
	// TODO the color doesn't work atm
	fragColor = vec4(vColor, 1.);
}
