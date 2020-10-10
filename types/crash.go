package types

// SerializableException defines an exception that follows this format, i.e. the one in https://github.com/soygul/NBug/blob/master/NBug/Core/Util/Serialization/SerializableException.cs
type SerializableException struct {
	Message    string
	Source     string
	StackTrace string
	TargetSite string
	Type       string
}
