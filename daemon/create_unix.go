// +build !windows

package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	containertypes "github.com/docker/docker/api/types/container"
	mounttypes "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/container"
	"github.com/docker/docker/oci"
	"github.com/docker/docker/pkg/stringid"
	volumeopts "github.com/docker/docker/volume/service/opts"
	"github.com/opencontainers/selinux/go-selinux/label"
	"github.com/sirupsen/logrus"
	u "github.com/YesZhen/superlog_go"
)

// createContainerOSSpecificSettings performs host-OS specific container create functionality
func (daemon *Daemon) createContainerOSSpecificSettings(container *container.Container, config *containertypes.Config, hostConfig *containertypes.HostConfig) error {
	defer u.LogEnd(u.LogBegin("createContainerOSSpecificSettings"))
	d, t := u.LogBegin("Mount")
	if err := daemon.Mount(container); err != nil {
		return err
	}
	u.LogEnd(d, t)
	defer daemon.Unmount(container)

//	d, t = u.LogBegin("RootPair")
	rootIDs := daemon.idMapping.RootPair()
//	u.LogEnd(d, t)
//	d, t = u.LogBegin("SetupWorkingDirectory")
	if err := container.SetupWorkingDirectory(rootIDs); err != nil {
		return err
	}
//	u.LogEnd(d, t)

//	d, t = u.LogBegin("if 1")
	// Set the default masked and readonly paths with regard to the host config options if they are not set.
	if hostConfig.MaskedPaths == nil && !hostConfig.Privileged {
		hostConfig.MaskedPaths = oci.DefaultSpec().Linux.MaskedPaths // Set it to the default if nil
		container.HostConfig.MaskedPaths = hostConfig.MaskedPaths
	}
//	u.LogEnd(d, t)
//	d, t = u.LogBegin("if 2")
	if hostConfig.ReadonlyPaths == nil && !hostConfig.Privileged {
		hostConfig.ReadonlyPaths = oci.DefaultSpec().Linux.ReadonlyPaths // Set it to the default if nil
		container.HostConfig.ReadonlyPaths = hostConfig.ReadonlyPaths
	}
//	u.LogEnd(d, t)
	u.Info("Time 4")
	d, t = u.LogBegin("volumes.Create")
	for spec := range config.Volumes {
		name := stringid.GenerateRandomID()
		destination := filepath.Clean(spec)

		// Skip volumes for which we already have something mounted on that
		// destination because of a --volume-from.
		if container.HasMountFor(destination) {
			logrus.WithField("container", container.ID).WithField("destination", spec).Debug("mountpoint already exists, skipping anonymous volume")
			// Not an error, this could easily have come from the image config.
			continue
		}
		path, err := container.GetResourcePath(destination)
		if err != nil {
			return err
		}

		stat, err := os.Stat(path)
		if err == nil && !stat.IsDir() {
			return fmt.Errorf("cannot mount volume over existing file, file exists %s", path)
		}
		u.Info("Time 10")
		u.Infof("name is: %s", name)
		u.Infof("driver name is: %s", hostConfig.VolumeDriver)
		u.Infof("container id is: %s", container.ID)
		v, err := daemon.volumes.Create(context.TODO(), name, hostConfig.VolumeDriver, volumeopts.WithCreateReference(container.ID))

		u.Infof("v name is: %s", v.Name)
		u.Infof("v Driver is: %s", v.Driver)
		u.Infof("v Mountpoint is: %s", v.Mountpoint)
		for labelkey := range v.Labels {
			u.Infof("labelKey is %s, labelValue is %s", labelkey, v.Labels[labelkey])
		}
		u.Infof("v Scope is: %s", v.Scope)

		u.Info("Time 11")
		if err != nil {
			return err
		}

		if err := label.Relabel(v.Mountpoint, container.MountLabel, true); err != nil {
			return err
		}

		container.AddMountPointWithVolume(destination, &volumeWrapper{v: v, s: daemon.volumes}, true)
		u.Infof("destination is: %s", destination)
	}
	u.LogEnd(d, t)
	u.Info("Time 5")
	return daemon.populateVolumes(container)
}

// populateVolumes copies data from the container's rootfs into the volume for non-binds.
// this is only called when the container is created.
func (daemon *Daemon) populateVolumes(c *container.Container) error {
	for _, mnt := range c.MountPoints {
		if mnt.Volume == nil {
			continue
		}

		if mnt.Type != mounttypes.TypeVolume || !mnt.CopyData {
			continue
		}

		logrus.Debugf("copying image data from %s:%s, to %s", c.ID, mnt.Destination, mnt.Name)
		if err := c.CopyImagePathContent(mnt.Volume, mnt.Destination); err != nil {
			return err
		}
	}
	return nil
}
