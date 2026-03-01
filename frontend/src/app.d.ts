import type { SvelteComponent } from "svelte";
import type { Writable as _Writable } from "svelte/store";
import type { KeymapMode } from "./routes/keymap-service";
import { type PickerKeymapMode } from "$lib/state.svelte";
// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {

  namespace App {
    interface NvimPosition {
      row: number;
      col: number;
    }

    interface NvimToken {
      text: string;
      classes: string;
      highlight: number;
    }

    interface NvimRow {
      index: number;
      tokens: NvimToken[]
      markdownOpts?: MarkdownOpts
    }

    interface TableRowOpts {
      tableId: number;
      rowType: 'header' | 'separator' | 'data';
      cells: NvimToken[][];
      alignments: string[];
    }

    interface ImageOpts {
      url: string;
      altText: string;
    }

    interface MarkdownOpts {
      quoteLevel: number
      table?: TableRowOpts
      image?: ImageOpts
    }

    interface WindowContentRowMap {
      [key: number]: NvimRow[];
    }
    interface NvimContent {
      winId: number
      content: NvimToken[][]
    }

    interface NvimLayout {
      cols: string
      rows: string
      activeWindowId: number
      windows: NvimWindow[]
    }

    type Modifier = 'c' | 's' | 'a'
    interface NvimWindow {
      id: number;
      type: string;
      width: number;
      height: number;
      colStart: number;
      colEnd: number;
      rowStart: number;
      rowEnd: number;
      lineNumbers: boolean;
      relativeLineNumbers: boolean;
      floatingWindows: FloatingWindow[];
      filetype: string;
      filepath: string;
      mode: VimMode;
      cursor: { row: number, col: number }
    }

    interface CmdLine {
      visible: boolean;
      firstc?: string;
      prompt?: string;
      content?: string;
      indent?: number;
      pos?: number;
    }
    interface Toast {
      level: NotificationLevel;
      text: string;
    }

    interface LiveGrepOpts {
      searchTerm: string;
      dir: string;
      include: string;
      exclude: string;
      caseSensitive: boolean;
      regex: boolean;
      matchWholeWord: boolean;
      totalResults: number;
    }
    interface RipgrepResult {
      Path: string;
      Matches: RipgrepMatch[]
    }

    interface RipgrepMatch {
      Id: string;
      FileName: string;
      AbsolutePath: string;
      MatchedLine: string;
      TextBeforeMatch: string;
      TextAfterMatch: string;
      MatchedText: string;
      ReplacementText: string;
      Row: number;
      Col: number;
      Html: string;
    }

    interface SearchResult {
      GroupedMatches: RipgrepResult[]
      PageIndex: number
      TotalPages: number
      TotalResults: number
      TotalFiles: number
    }

    interface PickerResult {
      filename: string;
      absolutePath: string;
      relativePath: string;
      icon: string;
      iconColor: string
      text?: string
      row?: number
      col?: number
    }
    interface NvimHighlight {
      id: number;
      fg: string;
      bg: string;
      bold: boolean;
      italic: boolean;
      underline: boolean;
      undercurl: boolean;
      strikethrough: boolean;
      reverse: boolean;
    }
    type VimMode = "normal" | "insert" | "visual" | "cmdline_normal" | "cmdline_insert"

    interface GridProps {
      content?: App.NvimRow[]
      decode?: (input: string) => string
      cursor?: { row: number, col: number }
      lineNumbers: boolean;
      relativeLineNumbers: boolean;
    }

    interface CursorMoveEvent {
      row: number;
      col: number;
      key: string;
      top_line: number;
      bottom_line: number;
    }

    type Writable<T> = _Writable<T>

    interface Picker {
      showEvent: string
      component: LegacyComponentType
      keymapMode: string
    }
    type KeymapMode = 'cmdline' | 'normal' | PickerKeymapMode
    interface Keymap {
      mods: Modifier[];
      mode: KeymapMode
      key: KeyboardEventKey;
      action: (e: KeyboardEvent) => void;
    }

    interface FloatingWindow {
      id: number;
      gridId: number;
      anchorWindow: number;
      anchor: string;
      row: number;
      col: number;
      width: number;
      height: number;
      zIndex: number;
      focusable: boolean;
      isPopup: boolean;
      isHex: boolean;
      filetype: string
    }

    type NamedKey =
      | 'Unidentified'
      | 'Alt'
      | 'AltGraph'
      | 'CapsLock'
      | 'Control'
      | 'Fn'
      | 'FnLock'
      | 'Meta'
      | 'NumLock'
      | 'ScrollLock'
      | 'Shift'
      | 'Symbol'
      | 'SymbolLock'
      | 'Enter'
      | 'Tab'
      | ' '
      | 'ArrowDown'
      | 'ArrowLeft'
      | 'ArrowRight'
      | 'ArrowUp'
      | 'End'
      | 'Home'
      | 'PageDown'
      | 'PageUp'
      | 'Backspace'
      | 'Clear'
      | 'Copy'
      | 'CrSel'
      | 'Cut'
      | 'Delete'
      | 'EraseEof'
      | 'ExSel'
      | 'Insert'
      | 'Paste'
      | 'Redo'
      | 'Undo'
      | 'Accept'
      | 'Again'
      | 'Attn'
      | 'Cancel'
      | 'ContextMenu'
      | 'Escape'
      | 'Execute'
      | 'Find'
      | 'Finish'
      | 'Help'
      | 'Pause'
      | 'Play'
      | 'Props'
      | 'Select'
      | 'ZoomIn'
      | 'ZoomOut'
      | 'BrightnessDown'
      | 'BrightnessUp'
      | 'Eject'
      | 'LogOff'
      | 'Power'
      | 'PowerOff'
      | 'PrintScreen'
      | 'Hibernate'
      | 'Standby'
      | 'WakeUp'
      | 'AllCandidates'
      | 'Alphanumeric'
      | 'CodeInput'
      | 'Compose'
      | 'Convert'
      | 'Dead'
      | 'FinalMode'
      | 'GroupFirst'
      | 'GroupLast'
      | 'GroupNext'
      | 'GroupPrevious'
      | 'ModeChange'
      | 'NextCandidate'
      | 'NonConvert'
      | 'PreviousCandidate'
      | 'Process'
      | 'SingleCandidate'
      | 'HangulMode'
      | 'HanjaMode'
      | 'JunjaMode'
      | 'Eisu'
      | 'Hankaku'
      | 'Hiragana'
      | 'HiraganaKatakana'
      | 'KanaMode'
      | 'KanjiMode'
      | 'Katakana'
      | 'Romaji'
      | 'Zenkaku'
      | 'ZenkakuHankaku'
      | 'F1'
      | 'F2'
      | 'F3'
      | 'F4'
      | 'F5'
      | 'F6'
      | 'F7'
      | 'F8'
      | 'F9'
      | 'F10'
      | 'F11'
      | 'F12'
      | 'F13'
      | 'F14'
      | 'F15'
      | 'F16'
      | 'F17'
      | 'F18'
      | 'F19'
      | 'F20'
      | 'F21'
      | 'F22'
      | 'F23'
      | 'F24'
      | 'F25'
      | 'F26'
      | 'F27'
      | 'F28'
      | 'F29'
      | 'F30'
      | 'F31'
      | 'F32'
      | 'F33'
      | 'F34'
      | 'F35'
      | 'Soft1'
      | 'Soft2'
      | 'Soft3'
      | 'Soft4'
      | 'ChannelDown'
      | 'ChannelUp'
      | 'Close'
      | 'MailForward'
      | 'MailReply'
      | 'MailSend'
      | 'MediaClose'
      | 'MediaFastForward'
      | 'MediaPause'
      | 'MediaPlay'
      | 'MediaPlayPause'
      | 'MediaRecord'
      | 'MediaRewind'
      | 'MediaStop'
      | 'MediaTrackNext'
      | 'MediaTrackPrevious'
      | 'New'
      | 'Open'
      | 'Print'
      | 'Save'
      | 'SpellCheck'
      | 'Key11'
      | 'Key12'
      | 'AudioBalanceLeft'
      | 'AudioBalanceRight'
      | 'AudioBassBoostDown'
      | 'AudioBassBoostToggle'
      | 'AudioBassBoostUp'
      | 'AudioFaderFront'
      | 'AudioFaderRear'
      | 'AudioSurroundModeNext'
      | 'AudioTrebleDown'
      | 'AudioTrebleUp'
      | 'AudioVolumeDown'
      | 'AudioVolumeMute'
      | 'AudioVolumeUp'
      | 'MicrophoneToggle'
      | 'MicrophoneVolumeDown'
      | 'MicrophoneVolumeMute'
      | 'MicrophoneVolumeUp'
      | 'TV'
      | 'TV3DMode'
      | 'TVAntennaCable'
      | 'TVAudioDescription'
      | 'TVAudioDescriptionMixDown'
      | 'TVAudioDescriptionMixUp'
      | 'TVContentsMenu'
      | 'TVDataService'
      | 'TVInput'
      | 'TVInputComponent1'
      | 'TVInputComponent2'
      | 'TVInputComposite1'
      | 'TVInputComposite2'
      | 'TVInputHDMI1'
      | 'TVInputHDMI2'
      | 'TVInputHDMI3'
      | 'TVInputHDMI4'
      | 'TVInputVGA1'
      | 'TVMediaContext'
      | 'TVNetwork'
      | 'TVNumberEntry'
      | 'TVPower'
      | 'TVRadioService'
      | 'TVSatellite'
      | 'TVSatelliteBS'
      | 'TVSatelliteCS'
      | 'TVSatelliteToggle'
      | 'TVTerrestrialAnalog'
      | 'TVTerrestrialDigital'
      | 'TVTimer'
      | 'AVRInput'
      | 'AVRPower'
      | 'ColorF0Red'
      | 'ColorF1Green'
      | 'ColorF2Yellow'
      | 'ColorF3Blue'
      | 'ColorF4Grey'
      | 'ColorF5Brown'
      | 'ClosedCaptionToggle'
      | 'Dimmer'
      | 'DisplaySwap'
      | 'DVR'
      | 'Exit'
      | 'FavoriteClear0'
      | 'FavoriteClear1'
      | 'FavoriteClear2'
      | 'FavoriteClear3'
      | 'FavoriteRecall0'
      | 'FavoriteRecall1'
      | 'FavoriteRecall2'
      | 'FavoriteRecall3'
      | 'FavoriteStore0'
      | 'FavoriteStore1'
      | 'FavoriteStore2'
      | 'FavoriteStore3'
      | 'Guide'
      | 'GuideNextDay'
      | 'GuidePreviousDay'
      | 'Info'
      | 'InstantReplay'
      | 'Link'
      | 'ListProgram'
      | 'LiveContent'
      | 'Lock'
      | 'MediaApps'
    type AlphanumericKey =
      | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j'
      | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't'
      | 'u' | 'v' | 'w' | 'x' | 'y' | 'z'
      | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J'
      | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T'
      | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'
      | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
      | '`' | '-' | '=' | '[' | ']' | '\\' | ';' | '\'' | ',' | '.' | '/'
      | '~' | '!' | '@' | '#' | '$' | '%' | '^' | '&' | '*' | '(' | ')'
      | '_' | '+' | '{' | '}' | '|' | ':' | '"' | '<' | '>' | '?';

    type KeyboardEventKey = NamedKey | AlphanumericKey;
  }
}

export { };
