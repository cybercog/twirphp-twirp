package gen

import (
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/twirphp/twirp/protoc-gen-twirp_php/internal/php"
	"github.com/twirphp/twirp/protoc-gen-twirp_php/templates/global"
	"github.com/twirphp/twirp/protoc-gen-twirp_php/templates/message"
	"github.com/twirphp/twirp/protoc-gen-twirp_php/templates/service"
)

const twirpVersion = "v8.1.0"

var (
	globalTemplates = template.Must(
		template.New("").Funcs(TxtFuncMap()).ParseFS(global.FS(), "*.php.tmpl"),
	)
	messageTemplates = template.Must(
		template.New("").Funcs(TxtFuncMap()).ParseFS(message.FS(), "*.php.tmpl"),
	)
	serviceTemplates = template.Must(
		template.New("").Funcs(TxtFuncMap()).ParseFS(service.FS(), "*.php.tmpl"),
	)
)

type serviceFileData struct {
	File           *protogen.File
	Service        *protogen.Service
	TwirpVersion   string
	TwirphpVersion string
	Version        string
}

type messageFileData struct {
	File    *protogen.File
	Message *protogen.Message
}

type globalFileData struct {
	Namespace string
	Version   string
}

// Generate is the main code generator.
func Generate(plugin *protogen.Plugin, version string) error {
	plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

	namespaces := map[string]bool{}

	println(plugin.Files)

	for _, file := range plugin.Files {
		for _, svc := range file.Services {
			// Namespaces are only relevant when there are services defined in them
			namespaces[php.Namespace(file)] = true

			for _, tpl := range serviceTemplates.Templates() {
				fileName := fmt.Sprintf(
					"%s/%s%s",
					php.Path(file),
					php.ServiceName(file, svc),
					strings.Replace(tpl.Name(), "_Service_", "", -1),
				)
				fileName = strings.TrimSuffix(fileName, ".tmpl")

				generatedFile := plugin.NewGeneratedFile(fileName, "")

				data := &serviceFileData{
					File:           file,
					Service:        svc,
					TwirpVersion:   twirpVersion,
					TwirphpVersion: version,
					Version:        version,
				}

				err := tpl.Execute(generatedFile, data)
				if err != nil {
					return err
				}
			}
		}

		for _, msg := range file.Messages {
			if msg.Desc.Name() != "GetEmployee" {
				continue
			}
			println(
				fmt.Sprintf(
					"desc.name=%s",
					msg.Desc.Name(),
				),
			)

			println("====")

			for _, tpl := range messageTemplates.Templates() {
				fileName := fmt.Sprintf(
					"%s/%s%s",
					"NeoTwirp/"+php.Path(file),
					msg.Desc.Name(),
					strings.Replace(tpl.Name(), "_Message_", "", -1),
				)
				fileName = strings.TrimSuffix(fileName, ".tmpl")

				generatedFile := plugin.NewGeneratedFile(fileName, "")

				data := &messageFileData{
					File:    file,
					Message: msg,
				}

				err := tpl.Execute(generatedFile, data)
				if err != nil {
					return err
				}
			}

			err := genMessageFile(plugin, file, msg)
			if err != nil {
				return err
			}
		}
	}

	for namespace := range namespaces {
		for _, tpl := range globalTemplates.Templates() {
			fileName := fmt.Sprintf("%s/%s", php.PathFromNamespace(namespace), tpl.Name())
			fileName = strings.TrimSuffix(fileName, ".tmpl")

			generatedFile := plugin.NewGeneratedFile(fileName, "")

			data := &globalFileData{
				Namespace: namespace,
				Version:   version,
			}

			err := tpl.Execute(generatedFile, data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func genMessageFile(
	plugin *protogen.Plugin,
	file *protogen.File,
	msg *protogen.Message,
) error {
	for _, innerMsg := range msg.Messages {
		for _, tpl := range messageTemplates.Templates() {
			fileName := fmt.Sprintf(
				"NeoTwirp/%s/%s%s",
				php.Path(file)+"/"+string(msg.Desc.Name()),
				innerMsg.Desc.Name(),
				strings.Replace(tpl.Name(), "_Message_", "", -1),
			)
			fileName = strings.TrimSuffix(fileName, ".tmpl")

			generatedFile := plugin.NewGeneratedFile(fileName, "")

			data := &messageFileData{
				File:    file,
				Message: innerMsg,
			}

			err := tpl.Execute(generatedFile, data)
			if err != nil {
				return err
			}
		}

		err := genMessageFile(plugin, file, innerMsg)
		if err != nil {
			return err
		}
	}
	return nil
}
