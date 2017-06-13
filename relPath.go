package dir

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RelSym(basepath, targetpath string) (string, error) {
	basepath, err := filepath.Abs(basepath)
	if err != nil {
		return "", fmt.Errorf("failed to abs and clean base: %v", err)
	}

	targetpath, err = filepath.Abs(targetpath)
	if err != nil {
		return "", fmt.Errorf("failed to abs and clean target: %v", err)
	}

	// #todo# cleanup
	// log.Printf("drive letter [%#v] [%#v] [%v}", basepath[0], targetpath[0],
	// 	basepath[0] == targetpath[0])

	if strings.ToLower(basepath)[0] != strings.ToLower(targetpath)[0] {
		return "", fmt.Errorf("windows drive letter differs")
	}

	// make sure you drop the first chunk no maatter what - either empty on *nix, or a drive letter on windows
	basepathChunks := strings.Split(basepath, string(os.PathSeparator))[1:]
	targetpathChunks := strings.Split(targetpath, string(os.PathSeparator))[1:]

	// log.Printf("\nchunks going in:\nprefix [%#v]\ntarget [%#v]\n", basepathChunks, targetpathChunks)

	relChunks, err := relSym([]string{}, basepathChunks, targetpathChunks)
	if err != nil {
		return "", err
	}

	return filepath.Join(relChunks...), nil
}

func relSym(basepath, prefix, target []string) ([]string, error) {
	// ##todo## cleanup
	log.Printf("start of child\nbasepath: %#v\nprefix: %#v\ntarget:%#v\n\n", basepath, prefix, target)

	// if you've cut off all of the prefix, return what you have
	if len(prefix) <= 0 {
		return target, nil
	}

	if len(target) <= 0 {
		return []string{}, errors.New("target is above prefix")
	}

	// ##todo## cleanup
	// log.Printf("[%s] [%s] [%v}", prefix[0], target[0], prefix[0] == target[0])
	if prefix[0] == target[0] {
		// call recursively, move one folder over
		return relSym(append(basepath, prefix[0]), prefix[1:], target[1:])
	}

	// strip off ..'s and move directories up as needed
	if prefix[0] == ".." {
		i := 1
		for prefix[i] == ".." {
			i++
		}
		return relSym(basepath[:len(basepath)-i], prefix[i:],
			append(basepath[len(basepath)-i:], target[:]...),
		)
	}
	if target[0] == ".." {
		i := 1
		for target[i] == ".." {
			i++
		}
		return relSym(basepath[:len(basepath)-i],
			append(basepath[len(basepath)-i:], prefix[:]...), target[i:])
	}

	// build the absolute version of each path
	basepathString := filepath.Join(append([]string{string(os.PathSeparator)}, basepath...)...)
	prefixPath := filepath.Join(basepathString, prefix[0])
	targetPath := filepath.Join(basepathString, prefix[0])

	// ##todo## cleanup
	// log.Printf("currently basepath [%s] prefix [%s] and target [%s]", basepathString, prefixPath, targetPath)
	// check both files to see if they're symlinks
	prefixSymInfo, err := os.Lstat(prefixPath)
	if err != nil {
		return append(basepath, prefix[0]), fmt.Errorf("failed to stat prefix: %v", err)
	}
	targetSymInfo, err := os.Lstat(targetPath)
	if err != nil {
		return append(basepath, target[0]), fmt.Errorf("failed to stat target: %v", err)
	}

	// read both symlinks
	var prefixAbs, targetAbs string
	var prefixRel, targetRel []string
	if prefixSymInfo.Mode()&os.ModeSymlink != 0 {
		prefixAbs, err = os.Readlink(prefixPath)
		if err != nil {
			return append(basepath, prefix[0]), fmt.Errorf("failed to readlink: %v", err)
		}
		if !strings.HasPrefix(prefixAbs, basepathString) {
			return []string{}, fmt.Errorf("base link missing from symlink: %s", prefixAbs)
		}
		prefixRel = strings.Split(strings.TrimPrefix(prefixAbs, basepathString), string(os.PathSeparator))
	}
	if targetSymInfo.Mode()&os.ModeSymlink != 0 {
		targetAbs, err = os.Readlink(targetPath)
		if err != nil {
			return append(basepath, target[0]), fmt.Errorf("failed to readlink: %v", err)
		}
		if !strings.HasPrefix(targetAbs, basepathString) {
			return []string{}, fmt.Errorf("base link missing from symlink: %s", targetAbs)
		}
		prefixRel = strings.Split(strings.TrimPrefix(targetAbs, basepathString), string(os.PathSeparator))
	}

	switch {
	case len(prefixRel) > 0 && prefixRel[0] == target[0]:
		// the prefix is a symlink that needs readlink to match
		prefix = append(prefixRel, prefix[1:]...)
		return relSym(append(basepath, prefix[0]), prefix[1:], target[1:])
	case len(targetRel) > 0 && prefix[0] == targetRel[0]:
		// the target is a symlink that needs readlink to match
		target = append(targetRel, target[1:]...)
		return relSym(append(basepath, prefix[0]), prefix[1:], target[1:])
	case len(prefixRel) > 0 && len(targetRel) > 0 && prefixRel[0] == targetRel[0]:
		// they match after you readlink both, so change both and go to the next step
		prefix = append(prefixRel, prefix[1:]...)
		target = append(targetRel, target[1:]...)
		return relSym(append(basepath, prefix[0]), prefix[1:], target[1:])
	default:
		return []string{}, fmt.Errorf("prefix [%s] is not a part of target [%s]",
			prefixPath, targetPath)
	}
}
