package runner

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
	fileutil "github.com/projectdiscovery/utils/file"
	updateutils "github.com/projectdiscovery/utils/update"
)

var (
	// retrieve home directory or fail
	homeDir = func() string {
		home, err := os.UserHomeDir()
		if err != nil {
			gologger.Fatal().Msgf("Failed to get user home directory: %s", err)
		}
		return home
	}()

	defaultConfigLocation = filepath.Join(homeDir, ".config/aix/config.yaml")
)

// Options contains the configuration options for tuning the enumeration process.
type Options struct {
	OpenaiApiKey       string              `yaml:"openai_api_key"`
	Prompt             string              `yaml:"prompt"`
	Gpt3               bool                `yaml:"gpt3"`
	Gpt4               bool                `yaml:"gpt4"`
	Model              string              `yaml:"model"`
	ListModels         bool                `yaml:"list_models"`
	Update             bool                `yaml:"update"`
	DisableUpdateCheck bool                `yaml:"disable_update_check"`
	Output             string              `yaml:"output"`
	Jsonl              bool                `yaml:"jsonl"`
	Verbose            bool                `yaml:"verbose"`
	Silent             bool                `yaml:"silent"`
	NoColor            bool                `yaml:"no_color"`
	Version            bool                `yaml:"version"`
	Stream             bool                `yaml:"stream"`
	TopP               float32             `yaml:"top_p"`
	Temperature        float32             `yaml:"temperature"`
	System             goflags.StringSlice `yaml:"system"`      // system message if any
	NoMarkdown         bool                `yaml:"no-markdown"` // render markdown message
}

// ParseOptions parses the command line flags provided by a user
func ParseOptions() *Options {
	options := &Options{}

	var temperature, topP string

	var promptArray goflags.StringSlice

	flagSet := goflags.NewFlagSet()

	flagSet.SetDescription(`AIx is a cli tool to interact with Large Language Model (LLM) APIs.`)

	flagSet.CreateGroup("input", "Input",
		flagSet.StringSliceVarP(&promptArray, "prompt", "p", nil, "prompt to query (input: stdin,string,file)", goflags.FileCommaSeparatedStringSliceOptions),
	)

	flagSet.CreateGroup("model", "Model",
		flagSet.BoolVarP(&options.Gpt3, "gpt3", "g3", true, "use GPT-3.5 model"),
		flagSet.BoolVarP(&options.Gpt4, "gpt4", "g4", false, "use GPT-4.0 model"),
		flagSet.StringVarP(&options.Model, "model", "m", "", "specify model to use (ex: gpt-4-0314)"),
		flagSet.BoolVarP(&options.ListModels, "list-models", "lm", false, "list available models"),
	)

	flagSet.CreateGroup("config", "Config",
		flagSet.StringVarP(&options.OpenaiApiKey, "openai-api-key", "ak", "", "openai api key token (input: string,file,env)"),
		flagSet.StringVarP(&temperature, "temperature", "t", "", "openai model temperature"),
		flagSet.StringVarP(&topP, "topp", "tp", "", "openai model top-p"),
		flagSet.StringSliceVarP(&options.System, "system-context", "sc", []string{}, "system message to send to the model (optional) (string,file)", goflags.FileNormalizedStringSliceOptions),
		flagSet.BoolVarP(&options.Stream, "stream", "s", false, "stream output to stdout (markdown rendering will be disabled)"),
	)

	flagSet.CreateGroup("update", "Update",
		flagSet.BoolVarP(&options.Update, "update", "up", false, "update aix to latest version"),
		flagSet.BoolVarP(&options.DisableUpdateCheck, "disable-update-check", "duc", false, "disable automatic aix update check"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.Output, "output", "o", "", "file to write output to"),
		flagSet.BoolVarP(&options.Jsonl, "jsonl", "j", false, "write output in json(line) format"),
		flagSet.BoolVarP(&options.Verbose, "verbose", "v", false, "verbose mode"),
		flagSet.BoolVar(&options.Silent, "silent", false, "display silent output"),
		flagSet.BoolVarP(&options.NoColor, "no-color", "nc", false, "disable colors in cli output"),
		flagSet.BoolVar(&options.Version, "version", false, "display project version"),
		flagSet.BoolVarP(&options.NoMarkdown, "no-markdown", "nm", false, "skip rendering markdown response"),
	)

	if err := flagSet.Parse(); err != nil {
		gologger.Fatal().Msgf("%s\n", err)
	}

	// Join prompt array into a single string
	options.Prompt = strings.Join(promptArray, "\n")

	if temperature != "" {
		val, err := strconv.ParseFloat(temperature, 32)
		if err == nil {
			options.Temperature = float32(val)
		}
	}
	if topP != "" {
		val, err := strconv.ParseFloat(topP, 32)
		if err == nil {
			options.TopP = float32(val)
		}
	}

	if fileutil.HasStdin() {
		bin, err := io.ReadAll(os.Stdin)
		if err != nil {
			gologger.Fatal().Msgf("couldn't read stdin: %s\n", err)
		}
		options.Prompt = string(bin)
	}

	if options.Prompt == "" {
		options.Prompt = strings.Join(flagSet.CommandLine.Args(), " ")
	}

	options.configureOutput()

	showBanner()

	if options.Version {
		gologger.Info().Msgf("Current Version: %s\n", version)
		os.Exit(0)
	}

	if options.OpenaiApiKey == "" {
		if key, exists := os.LookupEnv("OPENAI_API_KEY"); exists {
			options.OpenaiApiKey = key
		}
	}

	if options.OpenaiApiKey == "" {
		_ = options.loadConfigFrom(defaultConfigLocation)
	}

	if !options.DisableUpdateCheck {
		latestVersion, err := updateutils.GetToolVersionCallback("aix", version)()
		if err != nil {
			if options.Verbose {
				gologger.Error().Msgf("aix version check failed: %v", err.Error())
			}
		} else {
			gologger.Info().Msgf("Current aix version %v %v", version, updateutils.GetVersionDescription(version, latestVersion))
		}
	}

	return options
}

// configureOutput configures the output on the screen
func (options *Options) configureOutput() {
	// If the user desires verbose output, show verbose output
	if options.NoColor {
		gologger.DefaultLogger.SetFormatter(formatter.NewCLI(true))
	}
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	}
	if options.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
}

func (Options *Options) loadConfigFrom(location string) error {
	return fileutil.Unmarshal(fileutil.YAML, []byte(location), Options)
}
