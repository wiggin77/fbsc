package main

import "strings"

type loremParams interface {
	GetMaxWordsPerSentence() int
	GetMaxSentencesPerParagraph() int
	GetMaxParagraphs() int
}

func lorem(params loremParams) string {
	var sb = &strings.Builder{}

	maxParagraphs := params.GetMaxParagraphs()
	if maxParagraphs < 1 {
		maxParagraphs = 1
	}

	loops := pickRandomInt(1, maxParagraphs)
	for i := 0; i < loops; i++ {
		if i != 0 {
			sb.WriteString("\n")
		}
		addParagraph(sb, params)
	}
	return sb.String()
}

func addParagraph(sb *strings.Builder, params loremParams) {
	maxSentences := params.GetMaxSentencesPerParagraph()
	if maxSentences < 1 {
		maxSentences = 1
	}
	loops := pickRandomInt(1, maxSentences)
	for i := 0; i < loops; i++ {
		if i != 0 {
			sb.WriteString(" ")
		}
		addSentence(sb, params)
	}
}

func addSentence(sb *strings.Builder, params loremParams) {
	maxWordsPerSentence := params.GetMaxWordsPerSentence()
	if maxWordsPerSentence < 1 {
		maxWordsPerSentence = 1
	}

	loops := pickRandomInt(1, maxWordsPerSentence)
	for i := 0; i < loops; i++ {
		word := pickRandomString(loremDict)
		if i == 0 {
			word = strings.Title(word)
		} else {
			sb.WriteString(" ")
		}
		sb.WriteString(word)
	}
	sb.WriteString(".")
}

var loremDict = []string{
	"a",
	"ac",
	"accumsan",
	"ad",
	"adipiscing",
	"aenean",
	"aliquam",
	"aliquet",
	"amet",
	"ante",
	"aptent",
	"arcu",
	"at",
	"auctor",
	"augue",
	"bibendum",
	"blandit",
	"class",
	"commodo",
	"condimentum",
	"congue",
	"consectetur",
	"consequat",
	"conubia",
	"convallis",
	"cras",
	"curabitur",
	"cursus",
	"dapibus",
	"diam",
	"dictum",
	"dignissim",
	"dis",
	"dolor",
	"donec",
	"dui",
	"duis",
	"efficitur",
	"egestas",
	"eget",
	"eleifend",
	"elementum",
	"elit",
	"enim",
	"erat",
	"eros",
	"est",
	"et",
	"etiam",
	"eu",
	"euismod",
	"ex",
	"facilisi",
	"facilisis",
	"fames",
	"faucibus",
	"felis",
	"fermentum",
	"feugiat",
	"finibus",
	"fringilla",
	"fusce",
	"gravida",
	"habitant",
	"hendrerit",
	"himenaeos",
	"iaculis",
	"id",
	"imperdiet",
	"in",
	"inceptos",
	"integer",
	"interdum",
	"ipsum",
	"justo",
	"lacinia",
	"lacus",
	"laoreet",
	"lectus",
	"leo",
	"libero",
	"ligula",
	"litora",
	"lobortis",
	"lorem",
	"luctus",
	"maecenas",
	"magna",
	"magnis",
	"malesuada",
	"massa",
	"mattis",
	"mauris",
	"maximus",
	"metus",
	"mi",
	"molestie",
	"mollis",
	"montes",
	"morbi",
	"mus",
	"nam",
	"nascetur",
	"natoque",
	"nec",
	"neque",
	"netus",
	"nibh",
	"nisi",
	"nisl",
	"non",
	"nostra",
	"nulla",
	"nullam",
	"nunc",
	"odio",
	"orci",
	"ornare",
	"parturient",
	"pellentesque",
	"penatibus",
	"per",
	"pharetra",
	"phasellus",
	"placerat",
	"porta",
	"porttitor",
	"posuere",
	"potenti",
	"praesent",
	"pretium",
	"proin",
	"pulvinar",
	"purus",
	"quam",
	"quis",
	"quisque",
	"rhoncus",
	"ridiculus",
	"risus",
	"rutrum",
	"sagittis",
	"sapien",
	"scelerisque",
	"sed",
	"sem",
	"semper",
	"senectus",
	"sit",
	"sociosqu",
	"sodales",
	"sollicitudin",
	"suscipit",
	"suspendisse",
	"taciti",
	"tellus",
	"tempor",
	"tempus",
	"tincidunt",
	"torquent",
	"tortor",
	"tristique",
	"turpis",
	"ullamcorper",
	"ultrices",
	"ultricies",
	"urna",
	"ut",
	"varius",
	"vehicula",
	"vel",
	"velit",
	"venenatis",
	"vestibulum",
	"vitae",
	"vivamus",
	"viverra",
	"volutpat",
	"vulputate",
}
