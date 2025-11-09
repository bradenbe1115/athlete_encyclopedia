export UV_PYTHON_PREFERENCE := system

package_home ?= $(abspath .)
export VIRTUAL_ENV = $(package_home)/venv
prefix := cd $(package_home) &&
venv_bin := $(VIRTUAL_ENV)/bin

uv ?= uv
ifeq ($(UV_QUIET), 1)
uv += -q
endif

lock_file ?= requirements.lock.txt
runner ?= 

venv: 
	$(prefix) $(uv) venv --allow-existing $(VIRTUAL_ENV)

lock-deps: venv
	$(prefix) $(uv) pip compile --universal --no-header requirements.in -o $(lock_file)

deps: venv
	$(prefix) $(uv) pip sync -C editable_mode=compat $(lock_file)

test: deps
	$(prefix) $(venv_bin)/pytest -s .

run: deps
	$(prefix) $(venv_bin)/python ${runner}.py
