export CROS_CHECKOUT="${HOME}/chromiumos"

if [[ -d "${CROS_CHECKOUT}" ]]; then
    hash -d cros="${CROS_CHECKOUT}"
    hash -d ec="${CROS_CHECKOUT}/src/platform/ec"
    hash -d chromite="${CROS_CHECKOUT}/chromite"
    hash -d boardo="${CROS_CHECKOUT}/src/overlays"
    hash -d privateo="${CROS_CHECKOUT}/src/private-overlays"
    hash -d croso="${CROS_CHECKOUT}/src/third_party/chromiumos-overlay"
    hash -d portageo="${CROS_CHECKOUT}/src/third_party/portage-stable"
    hash -d p2="${CROS_CHECKOUT}/src/platform2"
    hash -d vboot="${CROS_CHECKOUT}/src/platform/vboot_reference"
    hash -d infraconf="${CROS_CHECKOUT}/infra/config"
    hash -d zmk="${CROS_CHECKOUT}/src/platform/ec/zephyr/zmake"
    hash -d mosys="${CROS_CHECKOUT}/src/platform/mosys"
    hash -d 3p="${CROS_CHECKOUT}/src/third_party"
    hash -d flashrom="${CROS_CHECKOUT}/src/third_party/flashrom"

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
        repo sync -n -j20 && repo rebase && repo --no-pager prune && repo sync -l -j$(nproc)
    }
fi
