// Copyright 2015 Matthew Collins
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package steven

import (
	"image"
	"math"

	"github.com/thinkofdeath/steven/render"
	"github.com/thinkofdeath/steven/render/builder"
	"github.com/thinkofdeath/steven/type/direction"
)

type processedModel struct {
	faces            []processedFace
	ambientOcclusion bool
	weight           int
}

type processedFace struct {
	cullFace        direction.Type
	facing          direction.Type
	vertices        []chunkVertex
	verticesTexture []render.TextureInfo
	indices         []int32
	shade           bool
	tintIndex       int
}

var faceRotation = []direction.Type{
	direction.North,
	direction.East,
	direction.South,
	direction.West,
}

var faceRotationX = []direction.Type{
	direction.North,
	direction.Down,
	direction.South,
	direction.Up,
}

func rotateDirection(val direction.Type, offset int, rots []direction.Type, invalid ...direction.Type) direction.Type {
	for _, d := range invalid {
		if d == val {
			return val
		}
	}
	var pos int
	for di, d := range rots {
		if d == val {
			pos = di
			break
		}
	}
	return rots[(pos+offset)%len(rots)]
}

func precomputeModel(bm *model) *processedModel {
	p := &processedModel{}
	p.ambientOcclusion = bm.ambientOcclusion
	p.weight = bm.weight
	for _, el := range bm.elements {
		for i, face := range el.faces {
			faceID := direction.Type(i)
			if face == nil {
				continue
			}
			pFace := processedFace{}
			cullFace := face.cullFace
			if bm.x > 0 {
				o := int(bm.x) / 90
				cullFace = rotateDirection(cullFace, o, faceRotationX, direction.East, direction.West, direction.Invalid)
				faceID = rotateDirection(faceID, o, faceRotationX, direction.East, direction.West, direction.Invalid)
			}
			if bm.y > 0 {
				o := int(bm.y) / 90
				cullFace = rotateDirection(cullFace, o, faceRotation, direction.Up, direction.Down, direction.Invalid)
				faceID = rotateDirection(faceID, o, faceRotation, direction.Up, direction.Down, direction.Invalid)
			}
			pFace.cullFace = cullFace
			pFace.facing = direction.Type(faceID)
			pFace.tintIndex = face.tintIndex
			pFace.shade = el.shade

			vert := faceVertices[i]
			tex := bm.lookupTexture(face.texture)
			rect := tex.Rect()

			ux1 := int16(face.uv[0] * float64(rect.Width))
			ux2 := int16(face.uv[2] * float64(rect.Width))
			uy1 := int16(face.uv[1] * float64(rect.Height))
			uy2 := int16(face.uv[3] * float64(rect.Height))

			tw, th := int16(rect.Width), int16(rect.Height)
			if face.rotation > 0 {
				x := ux1
				y := uy1
				w := ux2 - ux1
				h := uy2 - uy1
				switch face.rotation {
				case 90:
					uy2 = x + w
					ux1 = tw*16 - (y + h)
					ux2 = tw*16 - y
					uy1 = x
				case 180:
					uy1 = th*16 - (y + h)
					uy2 = th*16 - y
					ux1 = x + w
					ux2 = x
				case 270:
					uy2 = x
					uy1 = x + w
					ux2 = y + h
					ux1 = y
				}
			}

			var minX, minY, minZ = float32(math.Inf(1)), float32(math.Inf(1)), float32(math.Inf(1))
			var maxX, maxY, maxZ = float32(math.Inf(-1)), float32(math.Inf(-1)), float32(math.Inf(-1))

			for v := range vert.verts {
				pFace.verticesTexture = append(pFace.verticesTexture, tex)
				vert.verts[v].TX = uint16(rect.X)
				vert.verts[v].TY = uint16(rect.Y)
				vert.verts[v].TW = uint16(rect.Width)
				vert.verts[v].TH = uint16(rect.Height)
				vert.verts[v].TAtlas = int16(tex.Atlas())

				if vert.verts[v].X == 0 {
					vert.verts[v].X = float32(el.from[0] / 16.0)
				} else {
					vert.verts[v].X = float32(el.to[0] / 16.0)
				}
				if vert.verts[v].Y == 0 {
					vert.verts[v].Y = float32(el.from[1] / 16.0)
				} else {
					vert.verts[v].Y = float32(el.to[1] / 16.0)
				}
				if vert.verts[v].Z == 0 {
					vert.verts[v].Z = float32(el.from[2] / 16.0)
				} else {
					vert.verts[v].Z = float32(el.to[2] / 16.0)
				}

				if el.rotation != nil {
					r := el.rotation
					switch r.axis {
					case "y":
						rotY := -r.angle * (math.Pi / 180)
						c := math.Cos(rotY)
						s := math.Sin(rotY)
						x := float64(vert.verts[v].X) - (r.origin[0] / 16.0)
						z := float64(vert.verts[v].Z) - (r.origin[2] / 16.0)
						vert.verts[v].X = float32(r.origin[0]/16.0 + (x*c - z*s))
						vert.verts[v].Z = float32(r.origin[2]/16.0 + (z*c + x*s))
					case "x":
						rotX := r.angle * (math.Pi / 180)
						c := math.Cos(-rotX)
						s := math.Sin(-rotX)
						z := float64(vert.verts[v].Z) - (r.origin[2] / 16.0)
						y := float64(vert.verts[v].Y) - (r.origin[1] / 16.0)
						vert.verts[v].Z = float32(r.origin[2]/16.0 + (z*c - y*s))
						vert.verts[v].Y = float32(r.origin[1]/16.0 + (y*c + z*s))
					case "z":
						rotZ := -r.angle * (math.Pi / 180)
						c := math.Cos(-rotZ)
						s := math.Sin(-rotZ)
						x := float64(vert.verts[v].X) - (r.origin[0] / 16.0)
						y := float64(vert.verts[v].Y) - (r.origin[1] / 16.0)
						vert.verts[v].X = float32(r.origin[0]/16.0 + (x*c - y*s))
						vert.verts[v].Y = float32(r.origin[1]/16.0 + (y*c + x*s))
					}
				}

				if bm.x > 0 {
					rotX := bm.x * (math.Pi / 180)
					c := float32(math.Cos(rotX))
					s := float32(math.Sin(rotX))
					z := vert.verts[v].Z - 0.5
					y := vert.verts[v].Y - 0.5
					vert.verts[v].Z = 0.5 + (z*c - y*s)
					vert.verts[v].Y = 0.5 + (y*c + z*s)
				}

				if bm.y > 0 {
					rotY := bm.y * (math.Pi / 180)
					c := float32(math.Cos(rotY))
					s := float32(math.Sin(rotY))
					x := vert.verts[v].X - 0.5
					z := vert.verts[v].Z - 0.5
					vert.verts[v].X = 0.5 + (x*c - z*s)
					vert.verts[v].Z = 0.5 + (z*c + x*s)
				}

				if vert.verts[v].TOffsetX == 0 {
					vert.verts[v].TOffsetX = int16(ux1)
				} else {
					vert.verts[v].TOffsetX = int16(ux2)
				}
				if vert.verts[v].TOffsetY == 0 {
					vert.verts[v].TOffsetY = int16(uy1)
				} else {
					vert.verts[v].TOffsetY = int16(uy2)
				}

				if face.rotation > 0 {
					rotY := -float64(face.rotation) * (math.Pi / 180)
					c := int16(math.Cos(rotY))
					s := int16(math.Sin(rotY))
					x := vert.verts[v].TOffsetX - 8*tw
					y := vert.verts[v].TOffsetY - 8*th
					vert.verts[v].TOffsetX = 8*tw + int16(x*c-y*s)
					vert.verts[v].TOffsetY = 8*th + int16(y*c+x*s)
				}

				if bm.uvLock && bm.y > 0 &&
					(pFace.facing == direction.Up || pFace.facing == direction.Down) {
					rotY := float64(-bm.y) * (math.Pi / 180)
					c := int16(math.Cos(rotY))
					s := int16(math.Sin(rotY))
					x := vert.verts[v].TOffsetX - 8*16
					y := vert.verts[v].TOffsetY - 8*16
					vert.verts[v].TOffsetX = 8*16 + int16(x*c+y*s)
					vert.verts[v].TOffsetY = 8*16 + int16(y*c-x*s)
				}

				if bm.uvLock && bm.x > 0 &&
					(pFace.facing != direction.Up && pFace.facing != direction.Down) {
					rotY := float64(bm.x) * (math.Pi / 180)
					c := int16(math.Cos(rotY))
					s := int16(math.Sin(rotY))
					x := vert.verts[v].TOffsetX - 8*16
					y := vert.verts[v].TOffsetY - 8*16
					vert.verts[v].TOffsetX = 8*16 + int16(x*c+y*s)
					vert.verts[v].TOffsetY = 8*16 + int16(y*c-x*s)
				}

				if el.rotation != nil && el.rotation.rescale {
					if vert.verts[v].X < minX {
						minX = vert.verts[v].X
					} else if vert.verts[v].X > maxX {
						maxX = vert.verts[v].X
					}
					if vert.verts[v].Y < minY {
						minY = vert.verts[v].Y
					} else if vert.verts[v].Y > maxY {
						maxY = vert.verts[v].Y
					}
					if vert.verts[v].Z < minZ {
						minZ = vert.verts[v].Z
					} else if vert.verts[v].Z > maxZ {
						maxZ = vert.verts[v].Z
					}
				}
			}

			if el.rotation != nil && el.rotation.rescale {
				diffX := maxX - minX
				diffY := maxY - minY
				diffZ := maxZ - minZ
				for v := range vert.verts {
					vert.verts[v].X = (vert.verts[v].X - minX) / diffX
					vert.verts[v].Y = (vert.verts[v].Y - minY) / diffY
					vert.verts[v].Z = (vert.verts[v].Z - minZ) / diffZ
				}
			}

			pFace.vertices = vert.verts[:]
			pFace.indices = vert.indices[:]

			p.faces = append(p.faces, pFace)
		}
	}
	return p
}

func (p processedModel) Render(x, y, z int, bs *blocksSnapshot, buf *builder.Buffer, indices *int) {
	this := bs.block(x, y, z)
	for _, f := range p.faces {
		if f.cullFace != direction.Invalid {
			ox, oy, oz := f.cullFace.Offset()
			if b := bs.block(x+ox, y+oy, z+oz); b.ShouldCullAgainst() || b == this {
				continue
			}
		}

		var cr, cg, cb byte
		cr = 255
		cg = 255
		cb = 255
		switch f.tintIndex {
		case 0:
			if this.TintImage() != nil {
				cr, cg, cb = calculateBiome(bs, x, z, this.TintImage())
			} else {
				cr, cg, cb = this.TintColor()
			}
		}
		if f.facing == direction.West || f.facing == direction.East {
			cr = byte(float64(cr) * 0.8)
			cg = byte(float64(cg) * 0.8)
			cb = byte(float64(cb) * 0.8)
		}

		*indices += len(f.indices)

		for _, vert := range f.vertices {
			vert.R = cr
			vert.G = cg
			vert.B = cb

			vert.X += float32(x)
			vert.Y += float32(y)
			vert.Z += float32(z)

			vert.BlockLight, vert.SkyLight = calculateLight(
				bs,
				x, y, z,
				float64(vert.X),
				float64(vert.Y),
				float64(vert.Z),
				f.facing, p.ambientOcclusion, this.ForceShade(),
			)
			buildVertex(buf, &vert)
		}
	}
}

// Takes an average of the biome colors of the surrounding area
func calculateBiome(bs *blocksSnapshot, x, z int, img *image.NRGBA) (byte, byte, byte) {
	count := 0
	var r, g, b int
	for xx := -2; xx <= 2; xx++ {
		for zz := -2; zz <= 2; zz++ {
			biome := bs.biome(x+xx, z+zz)
			ix := biome.ColorIndex & 0xFF
			iy := biome.ColorIndex >> 8

			offset := img.PixOffset(ix, iy)
			r += int(img.Pix[offset])
			g += int(img.Pix[offset+1])
			b += int(img.Pix[offset+2])
			count++
		}
	}
	return byte(r / count), byte(g / count), byte(b / count)
}

func calculateLight(bs *blocksSnapshot, origX, origY, origZ int,
	x, y, z float64, face direction.Type, smooth, force bool) (uint16, uint16) {
	ox, oy, oz := face.Offset()
	if !bs.block(origX, origY, origZ).ShouldCullAgainst() {
		ox, oy, oz = 0, 0, 0
	}
	sblockLight := bs.blockLight(origX+ox, origY+oy, origZ+oz)
	sskyLight := bs.skyLight(origX+ox, origY+oy, origZ+oz)
	if !smooth {
		return uint16(sblockLight) * 4000, uint16(sskyLight) * 4000
	}
	blockLight := 0
	skyLight := 0
	count := 0

	dbl := int8(sblockLight) - 8
	if dbl < 0 {
		dbl = 0
	}
	sblockLight = byte(dbl)
	dsl := int8(sskyLight) - 8
	if dsl < 0 {
		dsl = 0
	}
	sskyLight = byte(dsl)

	ddx, ddy, ddz := face.Offset()
	dx, dy, dz := float64(ddx)*0.6, float64(ddy)*0.6, float64(ddz)*0.6
	for ox := -0.6; ox <= 0.0; ox += 0.6 {
		for oy := -0.6; oy <= 0.0; oy += 0.6 {
			for oz := -0.6; oz <= 0.0; oz += 0.6 {
				lx := round(x + ox + dx)
				ly := round(y + oy + dy)
				lz := round(z + oz + dz)
				bl := bs.blockLight(lx, ly, lz)
				sl := bs.skyLight(lx, ly, lz)
				if force && !bs.block(lx, ly, lz).Is(Blocks.Air) {
					bl = sblockLight
					sl = sskyLight
				}
				if sl == 0 && bl == 0 {
					bl = sblockLight
					sl = sskyLight
				}
				blockLight += int(bl)
				skyLight += int(sl)
				count++
			}
		}

	}

	return uint16((blockLight * 4000) / count), uint16((skyLight * 4000) / count)
}

func round(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

type faceDetails struct {
	indices [6]int32
	verts   [4]chunkVertex
}

// Precomputed face vertices
var faceVertices = [6]faceDetails{
	{ // Up
		indices: [6]int32{0, 1, 2, 3, 2, 1},
		verts: [4]chunkVertex{
			{X: 0, Y: 1, Z: 0, TOffsetX: 0, TOffsetY: 0},
			{X: 1, Y: 1, Z: 0, TOffsetX: 1, TOffsetY: 0},
			{X: 0, Y: 1, Z: 1, TOffsetX: 0, TOffsetY: 1},
			{X: 1, Y: 1, Z: 1, TOffsetX: 1, TOffsetY: 1},
		},
	},
	{ // Down
		indices: [6]int32{0, 1, 2, 3, 2, 1},
		verts: [4]chunkVertex{
			{X: 0, Y: 0, Z: 0, TOffsetX: 0, TOffsetY: 1},
			{X: 0, Y: 0, Z: 1, TOffsetX: 0, TOffsetY: 0},
			{X: 1, Y: 0, Z: 0, TOffsetX: 1, TOffsetY: 1},
			{X: 1, Y: 0, Z: 1, TOffsetX: 1, TOffsetY: 0},
		},
	},
	{ // North
		indices: [6]int32{0, 1, 2, 3, 2, 1},
		verts: [4]chunkVertex{
			{X: 0, Y: 0, Z: 0, TOffsetX: 1, TOffsetY: 1},
			{X: 1, Y: 0, Z: 0, TOffsetX: 0, TOffsetY: 1},
			{X: 0, Y: 1, Z: 0, TOffsetX: 1, TOffsetY: 0},
			{X: 1, Y: 1, Z: 0, TOffsetX: 0, TOffsetY: 0},
		},
	},
	{ // South
		indices: [6]int32{0, 1, 2, 3, 2, 1},
		verts: [4]chunkVertex{
			{X: 0, Y: 0, Z: 1, TOffsetX: 0, TOffsetY: 1},
			{X: 0, Y: 1, Z: 1, TOffsetX: 0, TOffsetY: 0},
			{X: 1, Y: 0, Z: 1, TOffsetX: 1, TOffsetY: 1},
			{X: 1, Y: 1, Z: 1, TOffsetX: 1, TOffsetY: 0},
		},
	},
	{ // West
		indices: [6]int32{0, 1, 2, 3, 2, 1},
		verts: [4]chunkVertex{
			{X: 0, Y: 0, Z: 0, TOffsetX: 0, TOffsetY: 1},
			{X: 0, Y: 1, Z: 0, TOffsetX: 0, TOffsetY: 0},
			{X: 0, Y: 0, Z: 1, TOffsetX: 1, TOffsetY: 1},
			{X: 0, Y: 1, Z: 1, TOffsetX: 1, TOffsetY: 0},
		},
	},
	{ // East
		indices: [6]int32{0, 1, 2, 3, 2, 1},
		verts: [4]chunkVertex{
			{X: 1, Y: 0, Z: 0, TOffsetX: 1, TOffsetY: 1},
			{X: 1, Y: 0, Z: 1, TOffsetX: 0, TOffsetY: 1},
			{X: 1, Y: 1, Z: 0, TOffsetX: 1, TOffsetY: 0},
			{X: 1, Y: 1, Z: 1, TOffsetX: 0, TOffsetY: 0},
		},
	},
}
