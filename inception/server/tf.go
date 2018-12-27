package main

// #cgo LDFLAGS: -Wl,-rpath -Wl,\$ORIGIN/lib
// import "C"

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/sdeoras/rpi-automation/grpc/inception"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

var (
	graph  *tf.Graph
	labels []string
)

type ClassifyResult struct {
	Filename string        `json:"filename"`
	Label    string        `json:"label"`
	Conf     int           `json:"conf"`
	Labels   []LabelResult `json:"labels"`
}

type LabelResult struct {
	Label       string  `json:"label"`
	Probability float32 `json:"probability"`
}

type ByProbability []LabelResult

func (a ByProbability) Len() int           { return len(a) }
func (a ByProbability) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByProbability) Less(i, j int) bool { return a[i].Probability > a[j].Probability }

func findBestLabels(image string, probabilities []float32) ClassifyResult {
	// Make a list of label/probability pairs
	var resultLabels []LabelResult
	for i, p := range probabilities {
		if i >= len(labels) {
			break
		}
		resultLabels = append(resultLabels, LabelResult{Label: labels[i], Probability: p})
	}
	// Sort by probability
	sort.Sort(ByProbability(resultLabels))
	// Return top 5 labels
	return ClassifyResult{Filename: image, Label: resultLabels[0].Label, Conf: int(resultLabels[0].Probability * 100), Labels: resultLabels[:5]}
}

func tfExec(image []byte) (*inception.Ack, error) {
	if err := loadGraph(); err != nil {
		return nil, err
	}

	ack := new(inception.Ack)
	ack.Label = make([]*inception.Label, 0, 0)

	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	tensor, err := makeTensorFromImage(bytes.NewBuffer(image), "jpg")
	if err != nil {
		return nil, err
	}

	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			graph.Operation("input").Output(0): tensor,
		},
		[]tf.Output{
			graph.Operation("output").Output(0),
		},
		nil)
	if err != nil {
		return nil, err
	}

	cr := findBestLabels("", output[0].Value().([][]float32)[0])

	for _, r := range cr.Labels {
		ack.Label = append(ack.Label, &inception.Label{Label: r.Label, Probability: int64(r.Probability * 100)})
	}
	ack.TopLabel = cr.Label
	ack.N = int64(cr.Conf)

	return ack, nil
}

func loadGraph() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, _ = filepath.Split(execPath)

	modelPath := filepath.Join(execPath, "model", "tensorflow_inception_graph.pb")
	model, err := ioutil.ReadFile(modelPath)
	if err != nil {
		return err
	}

	graph = tf.NewGraph()
	if err := graph.Import(model, ""); err != nil {
		return err
	}

	labelsPath := filepath.Join(execPath, "model", "imagenet_comp_graph_label_strings.txt")
	labelsFile, err := os.Open(labelsPath)
	if err != nil {
		return err
	}
	defer labelsFile.Close()

	scanner := bufio.NewScanner(labelsFile)
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func makeTensorFromImage(imageBuffer *bytes.Buffer, imageFormat string) (*tf.Tensor, error) {
	tensor, err := tf.NewTensor(imageBuffer.String())
	if err != nil {
		return nil, err
	}
	graph, input, output, err := makeTransformImageGraph(imageFormat)
	if err != nil {
		return nil, err
	}
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}

func makeTransformImageGraph(imageFormat string) (graph *tf.Graph, input, output tf.Output, err error) {
	const (
		H, W  = 224, 224
		Mean  = float32(117)
		Scale = float32(1)
	)
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	// Decode PNG or JPEG
	var decode tf.Output
	if imageFormat == "png" {
		decode = op.DecodePng(s, input, op.DecodePngChannels(3))
	} else {
		decode = op.DecodeJpeg(s, input, op.DecodeJpegChannels(3))
	}
	// Div and Sub perform (value-Mean)/Scale for each pixel
	output = op.Div(s,
		op.Sub(s,
			// Resize to 224x224 with bilinear interpolation
			op.ResizeBilinear(s,
				// Create a batch containing a single image
				op.ExpandDims(s,
					// Use decoded pixel values
					op.Cast(s, decode, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()
	return graph, input, output, err
}
