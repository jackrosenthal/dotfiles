export CROS_CHECKOUT="${HOME}/chromiumos"

if [[ -d "${CROS_CHECKOUT}" ]]; then
    hash -d cros="${CROS_CHECKOUT}"
    hash -d ec="${CROS_CHECKOUT}/src/platform/ec"
    hash -d chromite="${CROS_CHECKOUT}/chromite"
    hash -d boardo="${CROS_CHECKOUT}/src/overlays"
    hash -d privateo="${CROS_CHECKOUT}/src/private-overlays"
    hash -d croso="${CROS_CHECKOUT}/src/third_party/chromiumos-overlay"
    hash -d prortageo="${CROS_CHECKOUT}/src/third_party/portage-stable"
    hash -d p2="${CROS_CHECKOUT}/src/platform2"
    hash -d vboot="${CROS_CHECKOUT}/src/platform/vboot_reference"
    hash -d infraconf="${CROS_CHECKOUT}/infra/config"
    hash -d zmk="${CROS_CHECKOUT}/src/platform/ec/zephyr/zmake"
    hash -d mosys="${CROS_CHECKOUT}/src/platform/mosys"

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
fi
