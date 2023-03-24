package runner

import (
	"os"
	"path/filepath"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
	fileutil "github.com/projectdiscovery/utils/file"
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

	defaultConfigLocation = filepath.Join(homeDir, ".config/manx/config.yaml")
)

// Options contains the configuration options for tuning the enumeration process.
type Options struct {
	OpenaiApiKey       string `yaml:"openai_api_key"`
	Prompt             string `yaml:"prompt"`
	Gpt3               bool   `yaml:"gpt3"`
	Gpt4               bool   `yaml:"gpt4"`
	Update             bool   `yaml:"update"`
	DisableUpdateCheck bool   `yaml:"disable_update_check"`
	Output             string `yaml:"output"`
	Jsonl              bool   `yaml:"jsonl"`
	Verbose            bool   `yaml:"verbose"`
	Silent             bool   `yaml:"silent"`
	NoColor            bool   `yaml:"no_color"`
	Version            bool   `yaml:"version"`
}

// ParseOptions parses the command line flags provided by a user
func ParseOptions() *Options {
	options := &Options{}

	flagSet := goflags.NewFlagSet()

	flagSet.SetDescription(`manX is a golang based CLI tool to interact with Large Language Models (LLM) and Manual of everything.`)

	flagSet.CreateGroup("input", "Input",
		flagSet.StringVarP(&options.Prompt, "prompt", "p", "", "prompt to query (input: stdin,string,file)"),
	)

	flagSet.CreateGroup("model", "Model",
		flagSet.BoolVarP(&options.Gpt3, "gpt3", "g3", true, "use GPT-3.5 model (default)"),
		flagSet.BoolVarP(&options.Gpt4, "gpt4", "g4", false, "use GPT-4.0 model"),
	)

	flagSet.CreateGroup("config", "Config",
		flagSet.StringVarP(&options.OpenaiApiKey, "openai-api-key", "ak", "", "openai api key token (input: string,file,env)"),
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
	)

	if err := flagSet.Parse(); err != nil {
		gologger.Fatal().Msgf("%s\n", err)
	}

	if fileutil.HasStdin() {
		stdchan, err := fileutil.ReadFileWithReader(os.Stdin)
		if err != nil {
			gologger.Fatal().Msgf("couldn't read stdin: %s\n", err)
		}
		for prompt := range stdchan {
			options.Prompt = prompt
		}
	}

	options.configureOutput()

	showBanner()

	if options.Version {
		gologger.Info().Msgf("Current Version: %s\n", version)
		os.Exit(0)
	}

	if options.OpenaiApiKey == "" {
		_ = options.loadConfigFrom(defaultConfigLocation)
	}

	// if !options.DisableUpdateCheck {
	// 	latestVersion, err := updateutils.GetVersionCheckCallback("manx")()
	// 	if err != nil {
	// 		if options.Verbose {
	// 			gologger.Error().Msgf("manX version check failed: %v", err.Error())
	// 		}
	// 	} else {
	// 		gologger.Info().Msgf("Current manx version %v %v", version, updateutils.GetVersionDescription(version, latestVersion))
	// 	}
	// }

	return options
}

// configureOutput configures the output on the screen
func (options *Options) configureOutput() {
	// If the user desires verbose output, show verbose output
	if options.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
	if options.NoColor {
		gologger.DefaultLogger.SetFormatter(formatter.NewCLI(true))
	}
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	}
}

func (Options *Options) loadConfigFrom(location string) error {
	return fileutil.Unmarshal(fileutil.YAML, []byte(location), Options)
}

// just test for gh actions trigger
