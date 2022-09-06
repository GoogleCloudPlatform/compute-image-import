#!/usr/bin/env python3
# Copyright 2020 Google Inc. All Rights Reserved.
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
"""Perform inspection, and print results to stdout."""
import argparse
import base64
import os
import sys
import time

from boot_inspect import inspection
from compute_image_tools_proto import inspect_pb2
from google.protobuf import text_format
from google.protobuf.json_format import MessageToJson
import guestfs
import utils.diskutils as diskutils


def _daisy_kv(key: str, value: str):
  template = "Status: <serial-output key:'{key}' value:'{value}'>"
  return template.format(key=key, value=value)


def _output_daisy(results: inspect_pb2.InspectionResults):
  if results:
    print('Results: ')
    print(text_format.MessageToString(results))
    print(
        _daisy_kv(
            'inspect_pb',
            base64.standard_b64encode(results.SerializeToString()).decode()))
  print('Success: Done!')


def _output_human(results: inspect_pb2.InspectionResults):
  print(MessageToJson(results, indent=4))


def file_exists(file_path):
  if os.path.exists(file_path):
    print('File "{path}" is found.'.format(path=file_path))
    return True
  else:
    print('File "{path}" is not found.'.format(path=file_path))
    return False


def wait_for_device(device):
  success = False
  attached_disks = []
  for i in range(4):
    time.sleep(i * 10)

    attached_disks = diskutils.get_physical_drives()

    if device in attached_disks:
      success = True
      break
  if not success:
    msg = ("Input disk '{}' was not attached within "
           "the expected timeout").format(device)
    raise RuntimeError(msg)
  else:
    print("Input disk '{}' is attached successfully to "
          "the worker instance".format(device))


def main():
  format_options_and_help = {
    'json': 'JSON without newlines. Suitable for consumption by '
            'another program.',
    'human': 'Readable format that includes newlines and indents.',
    'daisy': 'Key-value format supported by Daisy\'s serial log collector.',
  }

  parser = argparse.ArgumentParser(
    description='Find boot-related properties of a disk.')
  parser.add_argument(
    '--format',
    choices=format_options_and_help.keys(),
    default='human',
    help=' '.join([
      '`%s`: %s' % (key, value)
      for key, value in format_options_and_help.items()
    ]))
  parser.add_argument(
    '--device',
    help='a block device (e.g. /dev/sdb).'
  )
  parser.add_argument(
    '--disk_file',
    help='The local path of a virtual disk file to inspect.'
  )

  args = parser.parse_args()
  if args.device is None and args.disk_file is None:
    print('either --disk_file or --device has to be specified')
    sys.exit(1)
  if args.device is not None and args.disk_file is not None:
    print('either --disk_file or --device has to be specified, but not both')
    sys.exit(1)

  disk_to_inspect = args.disk_file

  if args.device is not None:
    disk_to_inspect = args.device
    wait_for_device(disk_to_inspect)
  elif file_exists(args.disk_file) is False:
    sys.exit(1)

  results = inspect_pb2.InspectionResults()
  try:
    g = guestfs.GuestFS(python_return_dict=True)
    g.add_drive_opts(disk_to_inspect, readonly=1)
    g.launch()
  except BaseException as e:
    print('Failed to mount guest: ', e)
    results.ErrorWhen = inspect_pb2.InspectionResults.ErrorWhen.MOUNTING_GUEST
    globals()['_output_' + args.format](results)
    return

  try:
    print('Inspecting OS')
    results = inspection.inspect_device(g)
  except BaseException as e:
    print('Failed to inspect OS: ', e)
    results.ErrorWhen = inspect_pb2.InspectionResults.ErrorWhen.INSPECTING_OS

  try:
    boot_results = inspection.inspect_boot_loader(g, disk_to_inspect)
    results.bios_bootable = boot_results.bios_bootable
    results.uefi_bootable = boot_results.uefi_bootable
    results.root_fs = boot_results.root_fs
  except BaseException as e:
    print('Failed to inspect boot loader: ', e)
    results.ErrorWhen = \
        inspect_pb2.InspectionResults.ErrorWhen.INSPECTING_BOOTLOADER

  globals()['_output_' + args.format](results)


if __name__ == '__main__':
  main()
