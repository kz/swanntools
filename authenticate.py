# Attempt authentication with DVR4-1200

import argparse
import socket
import struct
import sys

INTENT_VALUES = "00000000000000000000010000000aXX000000292300000000001c010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
INTENT_RESPONSE_VALUES = "000000010000000aXX000000292300000000001c010000000100961200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
AUTH_VALUES = "000000000000000000000100000019YY0000000000000000000054000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
SUCCESSFUL_LOGIN_VALUES = "08 00 00 00 02 00 00 00"
FAILED_LOGIN_VALUES = "08 00 00 00 FF FF FF FF"


class Swann:
    def __init__(self, host, port, user, password, intent_value='7b'):
        self.sock = None
        self.get_socket(host, port)

        self.intent_message = None
        self.get_intent_message(intent_value)

        self.intent_response_message = None
        self.get_intent_response_message(intent_value)

        self.login_message = None
        self.get_login_message(user, password, intent_value)

    def get_socket(self, host, port):
        sys.stderr.write("Creating socket.\n")
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_TCP)
        self.sock.setsockopt(socket.SOL_SOCKET, socket.SO_RCVTIMEO, struct.pack('LL', 5, 0))
        self.sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
        self.sock.setsockopt(socket.SOL_SOCKET, socket.SO_LINGER, struct.pack('LL', False, 0))
        self.sock.connect((host, port))
        self.sock.setblocking(True)
        self.sock.settimeout(5)

    def get_intent_message(self, intent_value):
        hex_values = INTENT_VALUES
        hex_values = hex_values[:30] + intent_value + hex_values[32:]
        self.intent_message = bytearray.fromhex(hex_values)

    def get_intent_response_message(self, intent_value):
        hex_values = INTENT_RESPONSE_VALUES
        hex_values = hex_values[:16] + intent_value + hex_values[18:]
        self.intent_response_message = bytearray.fromhex(hex_values)

    def get_login_message(self, user, password, intent_value):
        hex_values = AUTH_VALUES

        hex_values = hex_values[:30] + format(int(intent_value, 16) + 1, 'x') + hex_values[32:]

        for i in range(0, len(user)):
            start_pos = 54 + 2 * i
            end_pos = 56 + 2 * i
            hex_values = hex_values[:start_pos] + format(ord(user[i]), 'x') + hex_values[end_pos:]

        for i in range(0, len(password)):
            start_pos = 118 + 2 * i
            end_pos = 120 + 2 * i
            hex_values = hex_values[:start_pos] + format(ord(password[i]), 'x') + hex_values[end_pos:]

        self.login_message = bytearray.fromhex(hex_values)

    def log_in(self):
        sys.stderr.write("Sending intent message.\n")
        self.sock.send(self.intent_message)
        sys.stderr.write("Intent message sent.\n")
        response = self.sock.recv(1000)
        sys.stderr.write("Intent response received.\n")

        if response == self.intent_response_message:
            sys.stderr.write("Intent response matches.\n")
        else:
            sys.stderr.write("Intent response does not match. Exiting.\n")
            exit()

        sys.stderr.write("Sending login message.\n")
        self.sock.send(self.login_message)
        sys.stderr.write("Login message sent.\n")

        response = self.sock.recv(8) # TODO: Currently receiving a zero-length response here. Perhaps the DVR is closing the socket early?
        self.sock.send(self.login_message)
        print(len(response))
        sys.stderr.write("Login response received.\n")
        # TODO: Compare response with bytearrays of the possible authentication responses, after comparing the response length


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Attempt to authenticate with DVR4-1200 via the media protocol')

    parser.add_argument('--host', type=str, required=True, help="Enter the DVR host")
    parser.add_argument('--port', type=int, required=True, help="Enter the media port")
    parser.add_argument('--user', type=str, required=True, default="admin", help="Specify a user name")
    parser.add_argument('--password', type=str, required=True, default="", help="Specify a password")
    parser.add_argument('--intent_value', type=str, required=False, default="7b", help="Specify an intent value")

    args = parser.parse_args()

    swann = Swann(args.host, args.port, args.user, args.password, args.intent_value)
    swann.log_in()
