FROM debian:bookworm

# Install sudo as it's required by install_deps.sh
RUN apt-get update && apt-get install -y sudo && rm -rf /var/lib/apt/lists/*

# Set up working directory
WORKDIR /rpi-image-gen

# Copy the rpi-image-gen source
COPY rpi-image-gen/ .

# Install dependencies
RUN ./install_deps.sh

# The config file will be mounted to /config.yaml
# The build output will be directed to /build (via -B /build)

ENTRYPOINT ["./rpi-image-gen", "build"]
CMD ["-c", "/config.yaml", "-B", "/build"]
