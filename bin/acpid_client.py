#!/usr/bin/env python3
import argparse
import enum
import pathlib
import re
import socket
import subprocess


registered_handlers = []


def handler(event_type_match):
    event_type_pattern = re.compile(event_type_match)

    def decorator(fcn):
        registered_handlers.append((event_type_pattern, fcn))
        return fcn

    return decorator


@handler("button/lid")
def handle_lid_change(event_type, acpi_name, lid_state):
    here = pathlib.Path(__file__).parent
    subprocess.run([here / "dockdet"], check=True)

    if lid_state == "open":
        subprocess.run(["xset", "dpms", "force", "on"], check=True)


def process_cmd(cmd):
    cmd = cmd.decode("utf-8")
    cmd_args = cmd.split()

    found_handler = False
    for event_type_pattern, event_handler in registered_handlers:
        if event_type_pattern.fullmatch(cmd_args[0]):
            found_handler = True
            print(f"Executing {event_handler.__name__} handler for event: {cmd}")
            event_handler(*cmd_args)

    if not found_handler:
        print(f"No handlers to match event: {cmd}")


def main():
    parser = argparse.ArgumentParser(description="acpid client")
    parser.add_argument(
        "-s", "--socket", default="/var/run/acpid.socket", help="acpid socket path"
    )

    opts = parser.parse_args()

    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    sock.connect(opts.socket)

    data = b""
    while True:
        data += sock.recv(2048)
        while True:
            endcmd = data.find(b"\n")
            if endcmd < 0:
                break
            process_cmd(data[:endcmd])
            data = data[endcmd + 1 :]


if __name__ == "__main__":
    main()
