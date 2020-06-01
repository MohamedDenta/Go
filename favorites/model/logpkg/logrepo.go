package logpkg
// set outputdata , process in log object
func (logObj LogStruct) WriteLog() {
	WriteOnlogFile(logObj)
}
