# Raspberry Pi Setup

The git submodule and subfolder rpi-image-gen is the raspberry project for image generation. in best case this setup can generate everything necessary without further operations on the raspi itself.

How to run it and more details: https://github.com/raspberrypi/rpi-image-gen/tree/master

The main folder should contain all the infos for using everything together to get the raspi running.

## Generating the Image with Docker

To build the image using Docker, follow these steps:

1. **Build the Docker container:**
   ```bash
   cd rpi
   docker build -t rpi-image-builder .
   ```

2. **Run the build:**
   Ensure you have a `build` directory created in the `rpi` folder to receive the output.
   ```bash
   mkdir -p build
   docker run --privileged \
     -v $(pwd)/config.yaml:/config.yaml \
     -v $(pwd)/build:/build \
     rpi-image-builder
   ```
   *Note: `--privileged` is required because `rpi-image-gen` uses namespaces and mounts during the build process.*

## Writing the Image to an SD Card

Once the build is complete, you will find the `.img` file in the `rpi/build` directory (specifically under a subdirectory created by the build process, e.g., `build/image-stargate-image/stargate-image.img`).

1. **Find your SD card device:**
   ```bash
   lsblk
   ```
   Look for your SD card (usually something like `/dev/sdb` or `/dev/mmcblk0`). **Make sure you identify the correct device, as `dd` will overwrite everything on it.**

2. **Write the image using `dd`:**
   Replace `/dev/sdX` with your actual device path and update the image path if necessary.
   ```bash
   sudo dd if=build/image-stargate-image/stargate-image.img of=/dev/sdX bs=4M status=progress conv=fsync
   ```

3. **Eject the card:**
   ```bash
   sudo eject /dev/sdX
   ```
