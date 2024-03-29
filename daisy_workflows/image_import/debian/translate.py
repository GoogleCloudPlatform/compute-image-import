#!/usr/bin/env python3
# Copyright 2017 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Translate the EL image on a GCE VM.

Parameters (retrieved from instance metadata):

debian_release: The version of the distro (stretch)
install_gce_packages: True if GCE agent and SDK should be installed
"""

import logging

import utils
import utils.diskutils as diskutils
from utils.guestfsprocess import run

google_cloud = '''
deb http://packages.cloud.google.com/apt cloud-sdk-{deb_release} main
deb http://packages.cloud.google.com/apt google-compute-engine-{deb_release}-stable main
deb http://packages.cloud.google.com/apt google-cloud-packages-archive-keyring-{deb_release} main
'''  # noqa: E501

interfaces = '''
source-directory /etc/network/interfaces.d
auto lo
iface lo inet loopback
auto eth0
iface eth0 inet dhcp
'''


def DistroSpecific(g):
  install_gce = utils.GetMetadataAttribute('install_gce_packages')
  deb_release = utils.GetMetadataAttribute('debian_release')

  if install_gce == 'true':
    logging.info('Installing GCE packages.')

    utils.update_apt(g)
    utils.install_apt_packages(g, 'gnupg')

    try:
      logging.debug('Adding Google Cloud apt-key.')
      cmd = ['wget', 'https://packages.cloud.google.com/apt/doc/apt-key.gpg',
             '-O', '/tmp/gce_key']
      run(g, cmd)
    except Exception as e:
      logging.debug('Failed to run wget command: ' + str(e))
      # check if curl is exist use it to add Google Cloud apt-key
      p = run(g, 'curl --version', raiseOnError=False)
      if p.code == 0:
        logging.debug('Trying to add Google Cloud apt-key with curl')
        cmd[0] = 'curl'
        cmd[2] = '-o'
        run(g, cmd)
      else:
        logging.debug('Installing wget')
        run(g, ['apt-get', 'install', '-y', 'wget'])
        run(g, cmd)
    run(g, ['apt-key', 'add', '/tmp/gce_key'])
    g.rm('/tmp/gce_key')
    g.write(
        '/etc/apt/sources.list.d/google-cloud.list',
        google_cloud.format(deb_release=deb_release))
    # Remove Azure agent.
    try:
      run(g, ['apt-get', 'remove', '-y', '-f', 'waagent', 'walinuxagent'])
    except Exception as e:
      logging.debug(str(e))
      logging.warn('Could not uninstall Azure agent. Continuing anyway.')

    utils.update_apt(g)
    pkgs = ['google-cloud-packages-archive-keyring', 'google-compute-engine']
    # Debian 8 differences:
    #   1. No NGE
    #   2. No Cloud SDK, since it requires Python 3.5+.
    #   3. No OS config agent.
    if deb_release == 'jessie':
      # Debian 8 doesn't support the new guest agent, so we need to install
      # the legacy Python version.
      pkgs += ['python-google-compute-engine',
               'python3-google-compute-engine']
      logging.info('Skipping installation of OS Config agent. '
                   'Requires Debian 9 or newer.')
    else:
      pkgs += ['google-cloud-sdk', 'google-osconfig-agent']
    utils.install_apt_packages(g, *pkgs)

  # Update grub config to log to console.
  run(g,
      ['sed', '-i""',
      r'/GRUB_CMDLINE_LINUX/s#"$# console=ttyS0,38400n8"#',
      '/etc/default/grub'])

  # Disable predictive network interface naming in 9+.
  if deb_release != 'jessie':
    run(g,
        ['sed', '-i',
        r's#^\(GRUB_CMDLINE_LINUX=".*\)"$#\1 net.ifnames=0 biosdevname=0"#',
        '/etc/default/grub'])

  run(g, ['update-grub2'])

  # Reset network for DHCP.
  logging.info('Resetting network to DHCP for eth0.')
  g.write('/etc/network/interfaces', interfaces)


def main():

  attached_disks = diskutils.get_physical_drives()

  # remove the boot disk of the worker instance
  attached_disks.remove('/dev/sda')

  g = diskutils.MountDisks(attached_disks)
  DistroSpecific(g)
  utils.CommonRoutines(g)
  diskutils.UnmountDisk(g)


if __name__ == '__main__':
  utils.RunTranslate(main)
