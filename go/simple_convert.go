package main

import (
	"fmt"
	"encoding/binary"

	"github.com/phil-mansfield/guppy/lib/compress"
	"github.com/phil-mansfield/guppy/lib/particles"
	"github.com/phil-mansfield/guppy/lib/snapio"
)

const (
	SnapMin = 100
	SnapMax = 100
	XDelta = 2.4e-3
	VDelta = 1.0
)

var (
	Span = [3]int{ 128, 128, 128 }
	FileMin = [3]int{ 0, 0, 0 }
	FileMax = [3]int{ 0, 0, 2 }
	GadgetTypes = []string{"v32", "v32", "u32"}
	GadgetNames = []string{"x", "v", "id"}
	Order = binary.LittleEndian
)

func InName(snap, fx, fy, fz int) string {
	switch fz {
	case 0: fx, fy, fz = 0, 0, 0
	case 1: fx, fy, fz = 3, 3, 3
	case 2: fx, fy, fz = 1, 2, 5
	}
	return fmt.Sprintf("/home/phil/code/src/github.com/phil-mansfield/" + 
		"guppy/large_test_data/L125_sheet%d%d%d_snap_%03d.gadget2.dat",
		fx, fy, fz, snap)
}

func OutName(snap, fx, fy, fz int) string {
	switch fz {
	case 0: fx, fy, fz = 0, 0, 0
	case 1: fx, fy, fz = 3, 3, 3
	case 2: fx, fy, fz = 1, 2, 5
	}
	return fmt.Sprintf("/home/phil/code/src/github.com/phil-mansfield/" + 
		"guppy/lib/compress/test_files/large_test_%d%d%d_snap%03d.gup",
		fx, fy, fz, snap)
}

func Names() (in, out []string) {
	for snap := SnapMin; snap <= SnapMax; snap++ {
		for fz := FileMin[2]; fz <= FileMax[2]; fz++ {
			for fy := FileMin[1]; fy <= FileMin[1]; fy++ {
				for fx := FileMin[0]; fx <= FileMin[0]; fx++{
					in = append(in, InName(snap, fx, fy, fz))
					out = append(out, OutName(snap, fx, fy, fz))
				}
			}
		}
	}

	return in, out
}

type Input struct {
	File snapio.File
	Header snapio.Header
	Buffer *snapio.Buffer
}

func InputBuffers(input0 string) *Input {
	var err error
	input := &Input{ }

	input.File, err = snapio.NewGadget2Cosmological(
		input0, GadgetNames, GadgetTypes, Order,
	)
	if err != nil { panic(err.Error()) }

	input.Header, err = input.File.ReadHeader()
	if err != nil { panic(err.Error()) }

	input.Buffer, err = snapio.NewBuffer(input.Header)
	if err != nil { panic(err.Error()) }

	return input
}

type Output struct {
	Buffer *compress.Buffer
	Methods map[string]compress.Method
	Data []float32
	B []byte
	Writer *compress.Writer
}

func OutputBuffers() *Output {
	methodMap := map[string]compress.Method{
		"x[0]": compress.NewLagrangianDelta(Span, XDelta),
		"x[1]": compress.NewLagrangianDelta(Span, XDelta),
		"x[2]": compress.NewLagrangianDelta(Span, XDelta),
		"v[0]": compress.NewLagrangianDelta(Span, VDelta),
		"v[1]": compress.NewLagrangianDelta(Span, VDelta),
		"v[2]": compress.NewLagrangianDelta(Span, VDelta),
	}
	return &Output{
		compress.NewBuffer(0),
		methodMap,
		make([]float32, Span[0]*Span[1]*Span[2]),
		[]byte{ },
		nil,
	}
}

func ReadVec(varName string, input *Input) [][3]float32 {
	err := input.File.Read(varName, input.Buffer)
	if err != nil { panic(err.Error()) }

	genericData, err := input.Buffer.Get(varName)
	if err != nil { panic(err.Error()) }

	vec, ok := genericData.([][3]float32)
	if !ok { panic("Impossible") }

	return vec
}

func WriteVec(varName string, vec [][3]float32, dim int, output *Output) {
	for i := range vec {
		output.Data[i] = vec[i][dim]
	}

	componentName := fmt.Sprintf("%s[%d]", varName, dim)
	component := particles.NewFloat32(componentName, output.Data)

	err := output.Writer.AddField(component, output.Methods[componentName])
	if err != nil { panic(err.Error()) }
}

func Convert(inName, outName string, input *Input, output *Output) {
	var err error

	input.File, err = snapio.NewGadget2Cosmological(
		inName, GadgetNames, GadgetTypes, Order,
	)
	if err != nil { panic(err.Error()) }

	output.Writer = compress.NewWriter(outName, output.Buffer, output.B, Order)

	for _, varName := range input.Header.Names() {
		if varName == "id" { continue }

		vec := ReadVec(varName, input)

		for dim := 0; dim < 3; dim++ {
			WriteVec(varName, vec, dim, output)
		}
	}

	output.B, err = output.Writer.Flush()
	input.Buffer.Reset()
}

func main() {
	inNames, outNames := Names()

	input := InputBuffers(inNames[0])
	output := OutputBuffers()

	for i := range inNames {
		Convert(inNames[i], outNames[i], input, output)
	} 
}