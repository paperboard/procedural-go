package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func Render(m Model, view mgl32.Mat4, project mgl32.Mat4, cameraPosition mgl32.Vec3) {
	lightPos := mgl32.Vec3{cameraPosition.X(), cameraPosition.Y(), cameraPosition.Z()}

	program := m.Program

	program.Use()
	gl.UniformMatrix4fv(program.GetUniformLocation("view"), 1, false, &view[0])
	gl.UniformMatrix4fv(program.GetUniformLocation("project"), 1, false, &project[0])
	gl.UniformMatrix4fv(program.GetUniformLocation("model"), 1, false, &m.Transform[0])

	gl.BindVertexArray(m.VAO)

	gl.Uniform3f(program.GetUniformLocation("lightColor"), 1.0, 1.0, 1.0)
	gl.Uniform3f(program.GetUniformLocation("lightPos"), lightPos.X(), lightPos.Y(), lightPos.Z())
	gl.Uniform1i(program.GetUniformLocation("textureId"), int32(m.TextureID))

	gl.DrawElements(gl.TRIANGLES, m.NbTriangles, gl.UNSIGNED_INT, gl.Ptr(m.Indices))

	gl.BindVertexArray(0)
}
