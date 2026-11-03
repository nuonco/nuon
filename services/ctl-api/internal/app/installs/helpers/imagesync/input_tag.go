package imagesync

import "strings"

func UseBuildForTag(decision Decision, tagBuildID, deployedBuildID, installComponentID string, tagMatches bool) Decision {
	if !tagMatches || tagBuildID == "" || installComponentID == "" {
		return decision
	}
	switch decision.Skip {
	case SkipNotAnImage, SkipNotInAppConfig, SkipNoInstallComponent:
		return decision
	}
	if tagBuildID == deployedBuildID {
		return Decision{Skip: SkipAlreadyCurrent}
	}
	return Decision{
		NeedsSync:          true,
		BuildID:            tagBuildID,
		InstallComponentID: installComponentID,
	}
}

func ImageTagMatches(resolvedTag, sourceRef, tag string) bool {
	if tag == "" {
		return false
	}
	if resolvedTag == tag {
		return true
	}
	return strings.HasSuffix(sourceRef, ":"+tag)
}
