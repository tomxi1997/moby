package containerd

import (
	"context"
	"reflect"
	"strings"

	"github.com/docker/docker/api/types/container"
	imagetype "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/image"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// MakeImageCache creates a stateful image cache.
func (i *ImageService) MakeImageCache(ctx context.Context, cacheFrom []string) (builder.ImageCache, error) {
	images := []*image.Image{}
	for _, c := range cacheFrom {
		im, err := i.GetImage(ctx, c, imagetype.GetImageOpts{})
		if err != nil {
			return nil, err
		}
		images = append(images, im)
	}
	return &imageCache{images: images, c: i}, nil
}

type imageCache struct {
	images []*image.Image
	c      *ImageService
}

func (ic *imageCache) GetCache(parentID string, cfg *container.Config, platform ocispec.Platform) (imageID string, err error) {
	ctx := context.TODO()

	if parentID == "" {
		// TODO handle "parentless" image cache lookups ("FROM scratch")
		return "", nil
	}

	parent, err := ic.c.GetImage(ctx, parentID, imagetype.GetImageOpts{Platform: &platform})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	for _, localCachedImage := range ic.images {
		if isMatch(localCachedImage, parent, cfg) {
			return localCachedImage.ID().String(), nil
		}
	}

	children, err := ic.c.Children(ctx, parent.ID())
	if err != nil {
		return "", err
	}

	for _, children := range children {
		childImage, err := ic.c.GetImage(ctx, children.String(), imagetype.GetImageOpts{Platform: &platform})
		if err != nil {
			if errdefs.IsNotFound(err) {
				continue
			}
			return "", err
		}

		if isMatch(childImage, parent, cfg) {
			return children.String(), nil
		}
	}

	return "", nil
}

// isMatch checks whether a given target can be used as cache for the given
// parent image/config combination.
// A target can only be an immediate child of the given parent image. For
// a parent image with `n` history entries, a valid target must have `n+1`
// entries and the extra entry must match the provided config
func isMatch(target, parent *image.Image, cfg *container.Config) bool {
	if target == nil || parent == nil || cfg == nil {
		return false
	}

	if len(target.History) != len(parent.History)+1 ||
		len(target.RootFS.DiffIDs) != len(parent.RootFS.DiffIDs)+1 {
		return false
	}

	for i := range parent.History {
		if !reflect.DeepEqual(parent.History[i], target.History[i]) {
			return false
		}
	}

	childCreatedBy := target.History[len(target.History)-1].CreatedBy
	return childCreatedBy == strings.Join(cfg.Cmd, " ")
}
