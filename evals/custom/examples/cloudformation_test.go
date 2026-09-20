package examples

import (
	"strings"
	"testing"
)

func TestCloudFormationAnnotationAndHeuristic(t *testing.T) {
	content := strings.Join([]string{
		"<!-- eval:cloudformation -->",
		"```yaml",
		"Resources:",
		"  Fn:",
		"    Type: AWS::Lambda::Function",
		"    Properties:",
		"      Runtime: nodejs22.x",
		"      Handler: index.handler",
		"```",
	}, "\n")
	blocks, err := extract("inline.md", content)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Annotation != AnnotationCloudFormation {
		t.Fatalf("cloudformation annotation not recognized: %+v", blocks)
	}
	if got := Classify(blocks[0].Documents()[0]); got != CategoryCloudFormation {
		t.Fatalf("annotated template classified as %q, want %q", got, CategoryCloudFormation)
	}

	if got := classifyContent(t, "yaml", `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  Fn:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: nodejs22.x
      Handler: index.handler
`, AnnotationNone); got != CategoryCloudFormation {
		t.Fatalf("CloudFormation heuristic = %q, want %q", got, CategoryCloudFormation)
	}
}

func TestValidateCloudFormationLambdaLayers(t *testing.T) {
	good := `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  Fn:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: nodejs22.x
      Handler: index.handler
      Layers:
        - arn:aws:lambda:eu-west-1:123456789012:layer:otel:1
  Extension:
    Type: Vendor::Observability::Extension
`
	if err := ValidateCloudFormation(good); err != nil {
		t.Fatalf("valid zip template rejected: %v", err)
	}

	intrinsicLayers := `Resources:
  Fn:
    Type: AWS::Serverless::Function
    Properties:
      Runtime: nodejs22.x
      Handler: index.handler
      Layers:
        Ref: OTelLayers
`
	if err := ValidateCloudFormation(intrinsicLayers); err != nil {
		t.Fatalf("intrinsic Layers expression rejected: %v", err)
	}

	imageWithLayers := `Resources:
  Fn:
    Type: AWS::Lambda::Function
    Properties:
      PackageType: Image
      Code:
        ImageUri: example.invalid/repo/image:TEST-1
      Layers:
        - arn:aws:lambda:eu-west-1:123456789012:layer:otel:1
`
	err := ValidateCloudFormation(imageWithLayers)
	if err == nil || !strings.Contains(err.Error(), "cannot use Layers") {
		t.Fatalf("image-with-layers error = %v, want cannot use Layers", err)
	}
}

func TestValidateDocumentCloudFormation(t *testing.T) {
	validator := &Validator{}
	block := &Block{
		File:       "inline.md",
		Line:       1,
		Tag:        "yaml",
		Annotation: AnnotationCloudFormation,
		Content: `Resources:
  Fn:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: nodejs22.x
      Handler: index.handler
`,
	}
	results := validator.validateDocument(block.Documents()[0])
	if len(results) != 1 || results[0].Status != StatusValidated {
		t.Fatalf("CloudFormation document did not validate: %+v", results)
	}
}
