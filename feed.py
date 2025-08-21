#!/usr/bin/env python3

import argparse
import os
import socket
import time
import collections
from datetime import datetime
from datetime import timedelta


class SignalLoader:
    def __init__(self):
        self.file_list = []
        self.curr_idx = -1

    def load_file(self, file_path):
        if file_path and os.path.isfile(file_path):
            _, file_ext = os.path.splitext(os.path.basename(file_path))
            if file_ext == '.dat' or file_ext == '.bvsp':
                self.file_list.append(os.path.abspath(file_path))

    def load_files(self, file_list):
        for file_path in file_list:
            self.load_file(file_path)

    def load_directory(self, dir_path):
        file_list = os.listdir(dir_path)
        for file_path in file_list:
            self.load_file(os.path.join(dir_path, file_path))

    def load_directories(self, dir_list):
        for dir_path in dir_list:
            self.load_directory(dir_path)

    def num_files(self):
        return len(self.file_list)

    def file_path_id(self, file_path):
        file_basename = os.path.basename(file_path)
        file_name, _ = os.path.splitext(file_basename)

        try:
            id = int(file_name, 10)
        except ValueError:
            id = 0

        return id

    def sort_files(self):
        self.file_list.sort(key=self.file_path_id)

    def has_next(self):
        if self.file_list:
            self.curr_idx += 1
            if self.curr_idx < len(self.file_list):
                return True
            else:
                return False
        else:
            return False

    def next_file(self):
        return self.file_list[self.curr_idx]

    def next(self):
        file_path = self.file_list[self.curr_idx]

        data = None
        try:
            fid = open(file_path, 'rb')
            data = fid.read()
            fid.close()
        except IOError as ioe:
            print(str(ioe))
            return None, file_path

        return data, file_path

    def rewind(self):
        self.curr_idx = -1

    def clean(self):
        self.file_list = []
        self.curr_idx = -1

    def is_empty(self):
        if self.file_list:
            return False
        else:
            return True


class SignalTransmitter:
    def __init__(self, ip="192.168.2.162", port=8000):
        self.ip = ip
        self.port = port
        self.sock = None
        self.connected = False

    def _connect(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((self.ip, self.port))

    def transmit(self, data):
        while not self.connected:
            print('Connecting to {}:{}'.format(self.ip, self.port))
            self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            try:
                self.sock.connect((self.ip, self.port))
            except ConnectionError:
                self.connected = False
                print('Connection error.')
                time.sleep(1)
            except TimeoutError:
                self.connected = False
                print('Connection timeout.')
                time.sleep(1)
            else:
                self.connected = True

        try:
            self.sock.sendall(data)
        except IOError:
            self.sock.close()
            self.connected = False
            return False

        return True

    def close(self):
        if self.sock:
            self.sock.close()
        self.connected = False

    def transmit_close(self, data):
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            sock.connect((self.ip, self.port))
            sock.sendall(data)
        except IOError:
            sock.close()
            return False
        sock.close()
        return True


def main(args):
    if not args.dir and not args.file:
        sig_dir = os.getcwd()
    else:
        sig_file = args.file
        sig_dir = args.dir
    ipaddr = args.ipaddr
    port = args.port
    interval_ms = args.interval
    tx_repeat = args.repeat

    sig_loader = SignalLoader()

    if sig_file:
        sig_loader.load_file(sig_file)
    if sig_dir:
        sig_loader.load_directory(sig_dir)

    sig_loader.sort_files()

    sig_transmitter = SignalTransmitter(ipaddr, port)

    num_pkts = 100
    time_buff = collections.deque(maxlen=num_pkts+1)
    time_buff.append(datetime.now())

    tx_repeated = 0
    while not tx_repeat or tx_repeated < tx_repeat:
        while sig_loader.has_next():
            data, filepath = sig_loader.next()
            sig_transmitter.transmit(data)
            time_buff.append(datetime.now())
            time_diff = time_buff[-1] - time_buff[0]
            time_per_pkt = time_diff / (len(time_buff)-1)

            print('Transmitted ({:.1f} ms/pkt): {}'.format(
                time_per_pkt.total_seconds()*1000, filepath))
            tx_repeated += 1
            time.sleep(interval_ms / 1000.0)
        sig_loader.rewind()


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Feed signals from file to PegaEngine.')
    parser.add_argument('-d', '--directory',
                        dest='dir',
                        help='load signal data from DIRECTORY, default the current working directory')
    parser.add_argument('-f', '--file',
                        dest='file',
                        help="load signal data from FILE, input can be repeated with Max 1024 files accepted")
    parser.add_argument('-i', '--ip',
                        dest="ipaddr",
                        default='127.0.0.1',
                        help='server IP address, default 127.0.0.1')
    parser.add_argument('-p', '--port',
                        dest="port",
                        type=int,
                        default=8000,
                        help='server TCP port, default 8000')
    parser.add_argument('-t', '--interval',
                        dest='interval',
                        type=int,
                        default=0,
                        help='transmit one packet of signal per INTERVAL milliseconds, default 0 ms')
    parser.add_argument('-r', '--repeat',
                        dest='repeat',
                        type=int,
                        default=0,
                        help='how many times to repeat the transmission, default 0 meaning infinite repeat')
    args = parser.parse_args()
    main(args)
