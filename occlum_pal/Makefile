SGX_SDK ?= /opt/intel/sgxsdk
SGX_MODE ?= HW

.PHONY: all build_src test clean

all: occlum_instance

occlum_instance: build_src
	@mkdir -p occlum_instance
	@cd occlum_instance && \
		occlum init && \
		rm -rf image && \
		copy_bom -f ../embedded_mode.yaml --root image --include-dir /opt/occlum/etc/template && \
		occlum build

build_src:
	@cd trusted/hello && \
		occlum-cargo build && \
		cd ..
	@$(MAKE) --no-print-directory -C driver

TOTAL_BYTES := 10000000000 # 10GB
test:
	LD_LIBRARY_PATH=driver/build:$(SGX_SDK)/sdk_libs RUST_BACKTRACE=1 \
		driver/build/driver $(TOTAL_BYTES)

clean:
	@cd trusted/hello && \
		occlum-cargo clean && \
		cd ..
	@$(MAKE) --no-print-directory -C driver clean
	@rm -rf .occlum occlum_instance
