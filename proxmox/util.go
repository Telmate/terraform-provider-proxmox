package proxmox

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rs/zerolog"
)

// given a string, return the appropriate zerolog level
func levelStringToZerologLevel(logLevel string) (zerolog.Level, error) {
	conversionMap := map[string]zerolog.Level{
		"panic": zerolog.PanicLevel,
		"fatal": zerolog.FatalLevel,
		"error": zerolog.ErrorLevel,
		"warn":  zerolog.WarnLevel,
		"info":  zerolog.InfoLevel,
		"debug": zerolog.DebugLevel,
		"trace": zerolog.TraceLevel,
	}

	foundResult, ok := conversionMap[logLevel]
	if !ok {
		return zerolog.Disabled, fmt.Errorf("Unable to find level %v", logLevel)
	}
	return foundResult, nil
}

// a global variable (but package scoped) to allow us to log stuff happening with style
// IMPORTANT:  this logger is created by the ConfigureLogger function.  Be sure that has run
// before using this logger otherwise you'll probably crash stuff.
var rootLogger zerolog.Logger

// a supporting global to keep track of our configured logLevels
// IMPORTANT:  this variable is set by the ConfigureLogger function.  Be sure that it has run.
var logLevels map[string]string

// Configure the debug logger for this provider.  The goal here is to enable selective amounts
// of output for targetted debugging without overwhelming with data from sources the user/developer
// doesn't care about.
//
// logLevels can be specifed as follows:
//   map[string]string
//
//   keys can be:
//    * '_root' - to affect the root logger
//    * '_capturelog' - (with any level set) to tell us to capture all message through the native log library
//    * '_default' - sets the default log level (if this is not set, the default is INFO)
//    (any other string) - the level to set that SubLogger to
//
//   Eventually we'll have a list of all subloggers that can be displayed/generated but for now, unfortuantely,
//   the code is the manual on that. I'll do my best to keep this doc string updated.
//
//   Known Subloggers:
//    * resource_vm_create - logs from the create function
//    * resource_vm_read  - logs from the read function
//
//   values can be one of "panic", "fatal", "error", "warn", "info", "debug", "trace".
//   these will be mapped out to the zerolog levels.  See the levelStringToZerologLevel function.
//
// logs will be written out to the logPath specified. An existing file at that path will be appended to.
// note that there are some information (like our redirection of the built-in log library) which will not
// follow the zerolog pattern and thus could mess with parsing.  This is annoying but something to fix in
// a future verison.
func ConfigureLogger(enableOutput bool, logPath string, inputLogLevels map[string]string) {

	// if we are not supposed to do anything here, then short circuit and do not set
	// anything up.
	if !enableOutput {
		rootLogger.Level(zerolog.Disabled)
		return
	}

	// update the global logLevels
	// I don't love globals, but feels like the right use here.
	logLevels = inputLogLevels

	// Create the log file if doesn't exist. And append to it if it already exists.
	// TODO log to stderr so at least terraform's TF_LOG can capture an issue if this file isn't created
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	// using a multi-writer here so we can easily add additional log destination (like a json file)
	// for now though using just the console writer because it makes pretty logs
	consoleWriter := zerolog.ConsoleWriter{Out: f, TimeFormat: time.RFC1123Z}
	multi := zerolog.MultiLevelWriter(consoleWriter)

	// create an init logger for logging just stuff before the root logger can get going
	// this has a hard coded set of information to ensure we can log stuff before the root logger is live
	initLogger := zerolog.New(multi).With().Timestamp().Caller().Logger().Level(zerolog.InfoLevel)

	// look to see if there is a default level we should be using
	defaultLevelString, ok := logLevels["_default"]
	if !ok {
		defaultLevelString = "info"
	}

	// set the log level using the default of INFO unless it is
	// overriden by the logLevels map by "_root" level
	rootLevelString, ok := logLevels["_root"]
	if !ok {
		rootLevelString = defaultLevelString
	}

	// translate the received log level into the zerolog Level type
	rootLevel, err := levelStringToZerologLevel(rootLevelString)
	if err != nil {
		initLogger.Info().Msgf("Received bad logLevel for _root logger: %v. Failing back to INFO level.", rootLevelString)
		rootLevel = zerolog.InfoLevel
	}

	// create the root logger
	// note there is no initialization here. we WANT this to be set to the global logger
	rootLogger = zerolog.New(multi).With().Timestamp().Caller().Logger().Level(rootLevel)

	// mirror Stdout to the debug log file as well
	// useful as we can debug the communication to/from the plugin and terraform
	origStdout := os.Stdout
	origStderr := os.Stderr
	mwriter := io.MultiWriter(f, origStdout)
	mwriterStderr := io.MultiWriter(f, origStderr)

	// get pipe reader and writer | writes to pipe writer come out pipe reader
	reader, writer, _ := os.Pipe()
	readerStderr, writerStderr, _ := os.Pipe()

	// replace stdout,stderr with pipe writer | all writes to stdout, stderr will go through pipe instead (fmt.print, log)
	os.Stdout = writer
	os.Stderr = writerStderr

	// look to see if we should capture all logs going through the native log library
	// this is mostly useful in this particular case to see logs from the proxmox api library.
	// just the presense of the _capturelog key (no matter the level set) is indication we should capture it
	_, ok = logLevels["_capturelog"]
	if ok {
		rootLogger.Info().Msg("Enabling the capture of log-library logs as ithe _capturelog flag was detected")
		log.SetOutput(f) // so we capture logs from any other dependencies not using logrus
	}

	//create channel to control exit | will block until all copies are finished
	communicateLogExit := make(chan bool)

	go func() {
		// copy all reads from pipe to multiwriter, which writes to stdout and file
		_, _ = io.Copy(mwriter, reader)
		// when r or w is closed copy will finish and true will be sent to channel
		communicateLogExit <- true
	}()

	go func() {
		// copy all reads from pipe to multiwriter, which writes to stdout and file
		_, _ = io.Copy(mwriterStderr, readerStderr)
		// when r or w is closed copy will finish and true will be sent to channel
		communicateLogExit <- true
	}()

	// yep this is a huge leak.. need to figure out a better way to close stuff down,
	// but for now, yolo!  we're just debugging.
	//
	//// function to be deferred in main until program exits
	//return func() {
	//	// close writer then block on exit channel | this will let mw finish writing before the program exits
	//	_ = w.Close()
	//	<-communicateLogExit
	//	// close file after all writes have finished
	//	_ = f.Close()
	//}

	rootLogger.Info().Msgf("Logging Started. Root Logger Set to level %v", rootLevel)
}

// Create a sublogger from the rootLogger
// This is helpful as it allows for custom logging level for each component/part of the system.
//
// The loggerName string is used to set the name of the logger in message outputs (as a key-val pair) but
// also as a way to know what we should set the logging level for this sublogger to (info/trace/warn/etc)
func CreateSubLogger(loggerName string) (zerolog.Logger, error) {

	// look to see if there is a default level we should be using
	defaultLevelString, ok := logLevels["_default"]
	if !ok {
		defaultLevelString = "info"
	}

	// set the log level using the default of INFO unless it is override by the logLevels map
	levelString, ok := logLevels[loggerName]
	if !ok {
		levelString = defaultLevelString
	}

	// translate the received log level into the zerolog Level type
	level, err := levelStringToZerologLevel(levelString)
	if err != nil {
		rootLogger.Info().Msgf("Received bad level %v when creating the %v sublogger. Failing back to INFO level.", levelString, loggerName)
		level = zerolog.InfoLevel
	}

	// create the logger
	thisLogger := rootLogger.With().Str("loggerName", loggerName).Logger().Level(level)
	return thisLogger, nil
}

func UpdateDeviceConfDefaults(
	activeDeviceConf pxapi.QemuDevice,
	defaultDeviceConf *schema.Set,
) *schema.Set {
	defaultDeviceConfMap := defaultDeviceConf.List()[0].(map[string]interface{})
	for key, _ := range defaultDeviceConfMap {
		if deviceConfigValue, ok := activeDeviceConf[key]; ok {
			defaultDeviceConfMap[key] = deviceConfigValue
			switch deviceConfigValue.(type) {
			case int:
				sValue := strconv.Itoa(deviceConfigValue.(int))
				bValue, err := strconv.ParseBool(sValue)
				if err == nil {
					defaultDeviceConfMap[key] = bValue
				}
			default:
				defaultDeviceConfMap[key] = deviceConfigValue
			}
		}
	}
	defaultDeviceConf.Remove(defaultDeviceConf.List()[0])
	defaultDeviceConf.Add(defaultDeviceConfMap)
	return defaultDeviceConf
}

func DevicesSetToMapWithoutId(devicesSet *schema.Set) pxapi.QemuDevices {
	devicesMap := pxapi.QemuDevices{}
	i := 1
	for _, set := range devicesSet.List() {
		setMap, isMap := set.(map[string]interface{})
		if isMap {
			// setMap["id"] = i
			devicesMap[i] = setMap
			i += 1
		}
	}
	return devicesMap
}

type KeyedDeviceMap map[interface{}]pxapi.QemuDevice

func DevicesSetToMapByKey(devicesSet *schema.Set, key string) KeyedDeviceMap {
	devicesMap := KeyedDeviceMap{}
	for i, set := range devicesSet.List() {
		setMap, isMap := set.(map[string]interface{})
		if isMap {
			if key != "" {
				devicesMap[setMap[key]] = setMap
			} else {
				devicesMap[i] = setMap
			}
		}
	}
	return devicesMap
}

func DeviceToMap(device pxapi.QemuDevice, key interface{}) KeyedDeviceMap {
	kdm := KeyedDeviceMap{}
	kdm[key] = device
	return kdm
}

func DevicesSetToDevices(devicesSet *schema.Set, key string) pxapi.QemuDevices {
	devicesMap := pxapi.QemuDevices{}
	for key, set := range DevicesSetToMapByKey(devicesSet, key) {
		devicesMap[key.(int)] = set
	}
	return devicesMap
}

func AddIds(configSet *schema.Set) *schema.Set {
	// add device config ids
	var i = 1
	for _, setConf := range configSet.List() {
		configSet.Remove(setConf)
		setConfMap := setConf.(map[string]interface{})
		setConfMap["id"] = i
		i += 1
		configSet.Add(setConfMap)
	}
	return configSet
}

func RemoveIds(configSet *schema.Set) *schema.Set {
	// remove device config ids
	for _, setConf := range configSet.List() {
		configSet.Remove(setConf)
		setConfMap := setConf.(map[string]interface{})
		delete(setConfMap, "id")
		configSet.Add(setConfMap)
	}
	return configSet
}
