package scr

import (
	"math"
	"time"

	"../cam"
	"../ctx"
	"../gfx"
	"../sky"
	"../ter"
	"../veg"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func RenderChunks(chunks []*ter.Chunk, camera *cam.FpsCamera, program *gfx.Program, textureContainer *ter.ChunkTextureContainer, dome *sky.Dome) {
	for _, chunk := range chunks {
		chunk.Model.Program = program
		RenderChunkModel(chunk.Model, camera, textureContainer, dome)
	}
}

func RenderSky(dome *sky.Dome, camera *cam.FpsCamera) {
	model := dome.Model
	model.Transform = mgl32.Translate3D(camera.Position().X(), 0, camera.Position().Z())
	program := model.Program
	program.Use()

	pvm := getPVM(model, camera)

	rotationStars := mgl32.Rotate3DX(-mgl32.DegToRad(float32(glfw.GetTime())))

	gl.UniformMatrix3fv(program.GetUniformLocation("u_rot_stars"), 1, false, &rotationStars[0])
	gl.UniformMatrix4fv(program.GetUniformLocation("u_pvm"), 1, false, &pvm[0])
	gl.Uniform3f(program.GetUniformLocation("u_sun_pos"), dome.SunPosition.X(), dome.SunPosition.Y(), dome.SunPosition.Z())
	gl.Uniform1i(program.GetUniformLocation("u_texture_tint"), int32(model.TextureID-gl.TEXTURE0))

	gl.BindVertexArray(model.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, model.Connectivity)
	gl.DrawElements(gl.TRIANGLES, model.NbTriangles*3, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

func RenderVegetation(gaia *veg.Gaia, camera *cam.FpsCamera, program *gfx.Program, dome *sky.Dome) {
	speed := 2.5
	amp := float32(2.5)
	angle := mgl32.DegToRad(amp * float32(math.Cos(speed*glfw.GetTime())))
	transformBasic := mgl32.Rotate3DX(angle).Mul3(mgl32.Rotate3DX(angle)).Mat4()

	gaia.InstanceGrass.Model.Program = program
	gaia.InstanceGrass.Model.Transform = transformBasic

	RenderInstances(gaia.InstanceGrass.Model, camera, dome, len(gaia.InstanceGrass.Transforms))

	for index, instanceTree := range gaia.InstanceTrees {
		for indexType, instanceTreeType := range instanceTree {
			var transform mgl32.Mat4
			if indexType == 1 {
				transform = transformBasic.Mul4(mgl32.Scale3D(5.0, 5.0, 5.0))
			} else {
				transform = transformBasic
			}
			instanceTreeType.BranchesModel.Program = program
			instanceTreeType.BranchesModel.TextureID = gl.TEXTURE1
			instanceTreeType.BranchesModel.Transform = transform
			RenderInstances(instanceTreeType.BranchesModel, camera, dome, len(instanceTreeType.Transforms))

			textureID := gl.TEXTURE10 + index%7
			instanceTreeType.LeavesModel.Program = program
			instanceTreeType.LeavesModel.TextureID = uint32(textureID)
			instanceTreeType.LeavesModel.Transform = transform
			RenderInstances(instanceTreeType.LeavesModel, camera, dome, len(instanceTreeType.Transforms))
		}
	}

}

func RenderChunkModel(m *gfx.Model, c *cam.FpsCamera, textureContainer *ter.ChunkTextureContainer, dome *sky.Dome) {
	m.Program.Use()
	initialiseUniforms(m, c, dome)
	gl.Uniform1i(m.Program.GetUniformLocation("time"), int32(time.Now().Unix()))
	setChunkTextureUniforms(m, textureContainer)
	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElements(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, nil)

	gl.BindVertexArray(0)
}

func RenderModel(m *gfx.Model, c *cam.FpsCamera, dome *sky.Dome) {
	m.Program.Use()
	initialiseUniforms(m, c, dome)

	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElements(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, nil)

	gl.BindVertexArray(0)
}

func setChunkTextureUniforms(m *gfx.Model, textureContainer *ter.ChunkTextureContainer) {
	gl.Uniform1i(m.Program.GetUniformLocation("dirtTexture"), int32(textureContainer.DirtID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("snowTexture"), int32(textureContainer.SnowID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("sandTexture"), int32(textureContainer.SandID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("rockTexture"), int32(textureContainer.RockID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("grassTexture"), int32(textureContainer.GrassID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("waterTexture"), int32(textureContainer.WaterID-gl.TEXTURE0))

}

func RenderInstances(m *gfx.Model, camera *cam.FpsCamera, dome *sky.Dome, nbrInstances int) {
	if nbrInstances == 0 {
		return
	}

	m.Program.Use()
	initialiseUniforms(m, camera, dome)
	gl.Uniform1i(m.Program.GetUniformLocation("u_nbr_instances"), int32(nbrInstances))
	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElementsInstanced(gl.TRIANGLES, m.NbTriangles, gl.UNSIGNED_INT, nil, int32(nbrInstances))
	gl.BindVertexArray(0)
}

func initialiseUniforms(m *gfx.Model, camera *cam.FpsCamera, dome *sky.Dome) {
	view := camera.GetTransform()
	project := mgl32.Perspective(mgl32.DegToRad(ctx.Fov),
		float32(ctx.Width())/float32(ctx.Height()), ctx.Near, ctx.Far)

	gl.Uniform1i(m.Program.GetUniformLocation("currentTexture"), int32(m.TextureID-gl.TEXTURE0))
	gl.Uniform1f(m.Program.GetUniformLocation("near"), ctx.Near)
	gl.Uniform1f(m.Program.GetUniformLocation("far"), ctx.Far)
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("view"), 1, false, &view[0])
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("project"), 1, false, &project[0])
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("model"), 1, false, &m.Transform[0])
	gl.Uniform3f(m.Program.GetUniformLocation("lightColor"), 1.0, 1.0, 1.0)
	gl.Uniform3f(m.Program.GetUniformLocation("lightPos"), dome.LightPosition.X(), dome.LightPosition.Y(), dome.LightPosition.Z())
	gl.Uniform1i(m.Program.GetUniformLocation("textureId"), int32(m.TextureID))

}

func getPVM(m *gfx.Model, camera *cam.FpsCamera) mgl32.Mat4 {
	view := camera.GetTransform()
	project := mgl32.Perspective(mgl32.DegToRad(ctx.Fov), float32(ctx.Width())/float32(ctx.Height()), ctx.Near, ctx.Far)
	model := m.Transform
	return project.Mul4(view).Mul4(model)
}
