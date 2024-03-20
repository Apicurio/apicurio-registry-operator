package env

import (
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"testing"
)

func TestJavaOptions(t *testing.T) {
	log := c.GetRootLogger(true)

	javaOptionsLegacy := "-Xms256m -Xmx1g -Dsimple=value -Ddots.dots.dots=true -Dspaces=\"one two three\" '-Dsingle quoted = true' \"-Ddouble quoted = true \""
	parsed, err := ParseShellArgs(javaOptionsLegacy)
	if err != nil {
		t.Fatal(err)
	}
	c.AssertEquals(t, map[string]string{
		"-Xms256m":         "",
		"-Xmx1g":           "",
		"-Dsimple":         "value",
		"-Ddots.dots.dots": "true",
		"-Dspaces":         "one two three",
		"-Dsingle quoted ": " true",
		"-Ddouble quoted ": " true ",
	}, parsed)

	javaOptions := "'-Dsingle quoted = false'          -Xms512          \n\n-Djust.quote=\"'\" -Djust.double.quote='\"'\n\n-Dspaces=\"one two three  four \""
	parsed2, err := ParseShellArgs(javaOptions)
	if err != nil {
		t.Fatal(err)
	}
	c.AssertEquals(t, map[string]string{
		"-Dsingle quoted ":    " false",
		"-Xms512":             "",
		"-Djust.quote":        "'",
		"-Djust.double.quote": "\"",
		"-Dspaces":            "one two three  four ",
	}, parsed2)

	cache := NewEnvCache(log)
	cache.Set(NewSimpleEnvCacheEntryBuilder(JAVA_OPTIONS_LEGACY, javaOptionsLegacy).SetPriority(PRIORITY_SPEC).Build())
	cache.Set(NewSimpleEnvCacheEntryBuilder(JAVA_OPTIONS, javaOptions).SetPriority(PRIORITY_SPEC).Build())

	expectedMerged := map[string]string{
		"-Xms256m":            "",
		"-Xmx1g":              "",
		"-Dsimple":            "value",
		"-Ddots.dots.dots":    "true",
		"-Dspaces":            "one two three  four ",
		"-Dsingle quoted ":    " false",
		"-Ddouble quoted ":    " true ",
		"-Xms512":             "",
		"-Djust.quote":        "'",
		"-Djust.double.quote": "\"",
	}

	parsed3, err := ParseJavaOptionsMap(cache)
	if err != nil {
		t.Fatal(err)
	}
	c.AssertEquals(t, expectedMerged, parsed3)

	SaveJavaOptionsMap(cache, parsed3)

	option, exists := cache.Get(JAVA_OPTIONS)
	if !exists {
		t.Fatal(JAVA_OPTIONS + " not found")
	}
	javaOptionsMerged := option.GetValue().Value
	parsed4, err := ParseShellArgs(javaOptionsMerged)
	if err != nil {
		t.Fatal(err)
	}
	c.AssertEquals(t, expectedMerged, parsed4)
}
