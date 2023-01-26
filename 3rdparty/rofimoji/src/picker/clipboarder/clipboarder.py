import subprocess
from subprocess import run

from ..abstractionhelper import is_installed, is_wayland
from ..typer.typer import Typer


class Clipboarder:
    @staticmethod
    def best_option(name: str = None) -> "Clipboarder":
        try:
            return next(clipboarder for clipboarder in Clipboarder.__subclasses__() if clipboarder.name() == name)()
        except StopIteration:
            try:
                return next(clipboarder for clipboarder in Clipboarder.__subclasses__() if clipboarder.supported())()
            except StopIteration:
                return Clipboarder()

    @staticmethod
    def supported() -> bool:
        pass

    @staticmethod
    def name() -> str:
        pass

    def copy_characters_to_clipboard(self, characters: str) -> None:
        raise NoClipboarderFoundException()

    def copy_paste_characters(self, characters: str, active_window: str, typer: Typer) -> None:
        raise NoClipboarderFoundException()


class XSelClipboarder(Clipboarder):
    @staticmethod
    def supported() -> bool:
        return not is_wayland() and is_installed("xsel")

    @staticmethod
    def name() -> str:
        return "xsel"

    def copy_characters_to_clipboard(self, characters: str) -> None:
        run(["xsel", "-i", "-b"], input=characters, encoding="utf-8")

    def copy_paste_characters(self, characters: str, active_window: str, typer: Typer) -> None:
        old_clipboard_content = run(args=["xsel", "-o", "-b"], capture_output=True).stdout
        old_primary_content = run(args=["xsel", "-o", "-p"], capture_output=True).stdout

        run(args=["xsel", "-i", "-b"], input=characters, encoding="utf-8")
        run(args=["xsel", "-i", "-p"], input=characters, encoding="utf-8")

        typer.insert_from_clipboard(active_window)

        run(args=["xsel", "-i", "-b"], input=old_clipboard_content)
        run(args=["xsel", "-i", "-p"], input=old_primary_content)


class XClipClipboarder(Clipboarder):
    @staticmethod
    def supported() -> bool:
        return not is_wayland() and is_installed("xclip")

    @staticmethod
    def name() -> str:
        return "xclip"

    def copy_characters_to_clipboard(self, characters: str) -> None:
        run(
            ["xclip", "-i", "-selection", "clipboard"],
            input=characters,
            encoding="utf-8",
            stderr=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
        )

    def copy_paste_characters(self, characters: str, active_window: str, typer: Typer) -> None:
        old_clipboard_content = run(args=["xclip", "-o", "-selection", "clipboard"], capture_output=True).stdout
        old_primary_content = run(args=["xclip", "-o", "-selection", "primary"], capture_output=True).stdout

        run(
            args=["xclip", "-i", "-selection", "clipboard"],
            input=characters,
            encoding="utf-8",
            stderr=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
        )
        run(
            args=["xclip", "-i", "-selection", "primary"],
            input=characters,
            encoding="utf-8",
            stderr=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
        )

        typer.insert_from_clipboard(active_window)

        run(
            args=["xclip", "-i", "-selection", "clipboard"],
            input=old_clipboard_content,
            stderr=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
        )
        run(
            args=["xclip", "-i", "-selection", "primary"],
            input=old_primary_content,
            stderr=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
        )


class WlClipboarder(Clipboarder):
    @staticmethod
    def supported() -> bool:
        return is_wayland() and is_installed("wl-copy")

    @staticmethod
    def name() -> str:
        return "wl-copy"

    def copy_characters_to_clipboard(self, characters: str) -> None:
        run(["wl-copy"], input=characters, encoding="utf-8")

    def copy_paste_characters(self, characters: str, active_window: str, typer: Typer) -> None:
        old_clipboard_content = run(args=["wl-paste"], capture_output=True).stdout
        old_primary_content = run(args=["wl-paste", "--primary"], capture_output=True).stdout

        run(args=["wl-copy"], input=characters, encoding="utf-8")
        run(args=["wl-copy", "--primary"], input=characters, encoding="utf-8")

        typer.insert_from_clipboard(active_window)

        run(args=["wl-copy"], input=old_clipboard_content)
        run(args=["wl-copy", "--primary"], input=old_primary_content)


class NoClipboarderFoundException(Exception):
    def __str__(self) -> str:
        return "Could not find a valid way to copy to clipboard. Please check the required dependencies."
