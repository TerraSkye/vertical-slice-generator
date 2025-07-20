package eventmodel

import "strings"

func AggregateTitle(title string) string {
	titleElements := ProcessTitle(title)
	base := Slugify(CapitalizeFirstCharacter(strings.ReplaceAll(titleElements, " ", "")))
	base = CapitalizeFirstCharacter(strings.ReplaceAll(base, "-", ""))
	if !strings.HasSuffix(title, "Aggregate") {
		base += "Aggregate"
	}
	return base
}

func CommandTitle(title string) string {
	titleElements := ProcessTitle(title)
	base := Slugify(CapitalizeFirstCharacter(strings.ReplaceAll(titleElements, " ", "")))
	base = strings.ReplaceAll(base, "-", "")
	if !strings.HasSuffix(title, "Command") {
		base += "Command"
	}
	return base
}

func FlowTitle(title string) string {
	title = strings.ReplaceAll(title, "flow:", "")
	titleElements := ProcessTitle(title)
	base := Slugify(CapitalizeFirstCharacter(strings.ReplaceAll(titleElements, " ", "")))
	base = strings.ReplaceAll(base, "-", "")
	if !strings.HasSuffix(title, "Flow") {
		base += "Flow"
	}
	return base
}

func SliceTitle(title string) string {
	title = strings.ReplaceAll(title, "slice:", "")
	title = strings.ReplaceAll(title, " ", "")
	return strings.ToLower(strings.ReplaceAll(Slugify(title), "-", ""))
}

func SliceSpecificClassTitle(slice, suffix string) string {
	slice = strings.ReplaceAll(slice, "-", "")
	parts := strings.Fields(slice)
	for i, part := range parts {
		parts[i] = CapitalizeFirstCharacter(SliceTitle(part))
	}
	return CapitalizeFirstCharacter(strings.Join(parts, "")) + suffix
}

func ProcessorTitle(title string) string {
	titleElements := ProcessTitle(title)
	base := Slugify(CapitalizeFirstCharacter(titleElements))
	base = strings.ReplaceAll(base, "-", "")
	return base + "Processor"
}

func RestResourceTitle(title string) string {
	titleElements := ProcessTitle(title)
	base := Slugify(CapitalizeFirstCharacter(titleElements))
	base = strings.ReplaceAll(base, "-", "")
	return base + "Resource"
}

func ReadModelTitle(title string) string {
	titleElements := ProcessTitle(title)
	base := Slugify(CapitalizeFirstCharacter(strings.ReplaceAll(titleElements, " ", "")))
	base = CapitalizeFirstCharacter(strings.ReplaceAll(base, "-", ""))
	return base + "ReadModel"
}

func EventTitle(title string) string {
	titleElements := ProcessTitle(title)
	base := Slugify(CapitalizeFirstCharacter(strings.ReplaceAll(titleElements, " ", "")))
	base = strings.ReplaceAll(base, "-", "")
	return base + "Event"
}

func ScreenTitle(title string) string {
	titleElements := ProcessTitle(title)
	return Slugify(CapitalizeFirstCharacter(strings.ReplaceAll(titleElements, " ", "")))
}

func PackageName(basePackage, contextPackage string, infrastructure bool) string {
	if infrastructure {
		return basePackage
	}
	if contextPackage != "" {
		return basePackage + "." + contextPackage
	}
	return basePackage
}

func PackageFolderName(basePackage, contextPackage string, infrastructure bool) string {
	return strings.ReplaceAll(PackageName(basePackage, contextPackage, infrastructure), ".", "/")
}

// shared logic
func ProcessTitle(title string) string {
	parts := strings.Fields(title)
	for i, part := range parts {
		clean := strings.ReplaceAll(part, "-", "")
		parts[i] = CapitalizeFirstCharacter(clean)
	}
	return strings.Join(parts, "")
}
