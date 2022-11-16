FROM nvidia/cuda:10.2-cudnn8-devel-ubuntu18.04

RUN apt-get update
RUN apt-get install -y python3 python3-pip
RUN python3 -m pip install --upgrade pip
RUN python3 -m pip install --upgrade Pillow
RUN pip3 install torch torchvision torchaudio

WORKDIR /work
