package tags

import (
	"sort"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/parse"
)

func removeDuplicates(tags *pveSDK.Tags) *pveSDK.Tags {
	if tags == nil || len(*tags) == 0 {
		return nil
	}
	tagMap := make(map[pveSDK.Tag]struct{})
	for _, tag := range *tags {
		tagMap[tag] = struct{}{}
	}
	uniqueTags := make(pveSDK.Tags, len(tagMap))
	var index uint
	for tag := range tagMap {
		uniqueTags[index] = tag
		index++
	}
	return &uniqueTags
}

func sortArray(tags *pveSDK.Tags) *pveSDK.Tags {
	if tags == nil || len(*tags) == 0 {
		return nil
	}
	sort.SliceStable(*tags, func(i, j int) bool {
		return (*tags)[i] < (*tags)[j]
	})
	return tags
}

func split(rawTags string) *pveSDK.Tags {
	return new(pveSDK.Tags(parse.DoubleDelimitedString(rawTags, func(s string) pveSDK.Tag {
		return pveSDK.Tag(s)
	})))
}

func toString(tags *pveSDK.Tags) (tagList string) {
	if tags == nil || len(*tags) == 0 {
		return ""
	}
	for _, tag := range *tags {
		tagList += ";" + string(tag)
	}
	return tagList[1:]
}
