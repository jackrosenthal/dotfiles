export CROS_CHECKOUT="${HOME}/chromiumos"

if [[ -d "${CROS_CHECKOUT}" ]]; then
    hash -d 3p="${CROS_CHECKOUT}/src/third_party"
    hash -d bazel="${CROS_CHECKOUT}/src/bazel"
    hash -d boardo="${CROS_CHECKOUT}/src/overlays"
    hash -d build="${CROS_CHECKOUT}/out/build"
    hash -d chromite="${CROS_CHECKOUT}/chromite"
    hash -d chroot="${CROS_CHECKOUT}/chroot"
    hash -d cros="${CROS_CHECKOUT}"
    hash -d croso="${CROS_CHECKOUT}/src/third_party/chromiumos-overlay"
    hash -d crosutils="${CROS_CHECKOUT}/src/scripts"
    hash -d ec="${CROS_CHECKOUT}/src/platform/ec"
    hash -d flashrom="${CROS_CHECKOUT}/src/third_party/flashrom"
    hash -d infraconf="${CROS_CHECKOUT}/infra/config"
    hash -d mosys="${CROS_CHECKOUT}/src/platform/mosys"
    hash -d out="${CROS_CHECKOUT}/out"
    hash -d p2="${CROS_CHECKOUT}/src/platform2"
    hash -d portageo="${CROS_CHECKOUT}/src/third_party/portage-stable"
    hash -d privateo="${CROS_CHECKOUT}/src/private-overlays"
    hash -d recipes="${CROS_CHECKOUT}/infra/recipes"
    hash -d vboot="${CROS_CHECKOUT}/src/platform/vboot_reference"
    hash -d zmk="${CROS_CHECKOUT}/src/platform/ec/zephyr/zmake"

    sdk() {
        local enter_args=()
        case "$(realpath "${PWD}")" in
            "${CROS_CHECKOUT}"* )
                enter_args+=(--working-dir=.)
                ;;
        esac
        TERM=xterm-256color "${CROS_CHECKOUT}/chromite/bin/cros_sdk" \
                            "${enter_args[@]}" -- "$@"
    }

    iuse() {
        flag="$1"
        shift
        cros query ebuilds -f "'${flag}' in iuse" "$@"
    }

    brduse() {
        flag="$1"
        shift
        cros query boards -f "'${flag}' in use_flags" "$@"
    }

    overlay() {
        cd "$(cros query overlays -f "name == '$1'")"
    }

    cros_sync() {
        repo sync -n -j20 --optimized-fetch && repo rebase --onto-manifest && repo --no-pager prune && repo sync -l -j$(nproc)
    }
fi
