package template

type commandTemplate struct {
	info         *GenerationInfo
	ignoreParams map[string][]string
	lenParams    map[string][]string
}

//func NewActivityTemplate(info *GenerationInfo) Template {
//	return &activityTemplate{
//		info: info,
//	}
//}
