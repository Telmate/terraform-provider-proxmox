package tags

import (
	"sort"
	"strings"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Returns an unordered list of unique tags
func RemoveDuplicates(tags *[]pxapi.Tag) *[]pxapi.Tag {
	if tags == nil || len(*tags) == 0 {
		return nil
	}
	tagMap := make(map[pxapi.Tag]struct{})
	for _, tag := range *tags {
		tagMap[tag] = struct{}{}
	}
	uniqueTags := make([]pxapi.Tag, len(tagMap))
	var index uint
	for tag := range tagMap {
		uniqueTags[index] = tag
		index++
	}
	return &uniqueTags
}

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf("expected a string, got: %s", i)
			}
			for _, e := range *Split(v) {
				if err := e.Validate(); err != nil {
					return diag.Errorf("tag validation failed: %s", err)
				}
			}
			return nil
		},
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return String(sortArray(RemoveDuplicates(Split(old)))) == String(sortArray(RemoveDuplicates(Split(new))))
		},
	}
}

func sortArray(tags *[]pxapi.Tag) *[]pxapi.Tag {
	if tags == nil || len(*tags) == 0 {
		return nil
	}
	sort.SliceStable(*tags, func(i, j int) bool {
		return (*tags)[i] < (*tags)[j]
	})
	return tags
}

func Split(rawTags string) *[]pxapi.Tag {
	tags := make([]pxapi.Tag, 0)
	if rawTags == "" {
		return &tags
	}
	tagArrays := strings.Split(rawTags, ";")
	for _, tag := range tagArrays {
		tagSubArrays := strings.Split(tag, ",")
		if len(tagSubArrays) > 1 {
			tmpTags := make([]pxapi.Tag, len(tagSubArrays))
			for i, e := range tagSubArrays {
				tmpTags[i] = pxapi.Tag(e)
			}
			tags = append(tags, tmpTags...)
		} else {
			tags = append(tags, pxapi.Tag(tag))
		}
	}
	return &tags
}

func String(tags *[]pxapi.Tag) (tagList string) {
	if tags == nil || len(*tags) == 0 {
		return ""
	}
	for _, tag := range *tags {
		tagList += ";" + string(tag)
	}
	return tagList[1:]
}
