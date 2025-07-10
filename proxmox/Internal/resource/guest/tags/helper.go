package tags

import (
	"sort"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
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
	tags := make(pveSDK.Tags, 0)
	if rawTags == "" {
		return &tags
	}
	tagIter := strings.SplitSeq(rawTags, ";")
	for tag := range tagIter {
		tagSubArrays := strings.Split(tag, ",")
		if len(tagSubArrays) > 1 {
			tmpTags := make([]pveSDK.Tag, len(tagSubArrays))
			for i, e := range tagSubArrays {
				tmpTags[i] = pveSDK.Tag(e)
			}
			tags = append(tags, tmpTags...)
		} else {
			tags = append(tags, pveSDK.Tag(tag))
		}
	}
	return &tags
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
