export interface SubstituteCommand {
  range: string
  search: string;
  replace: string
  flags: string;
  hasVeryMagic: boolean;
  hasGlobal: boolean;
  hasIgnoreCase: boolean;
  hasCaseSensitive: boolean;
  hasConfirm: boolean;
  searchPos?: number
  replacePos?: number
  isInSearchField: boolean;
  isInReplaceField: boolean;
  searchStart: number;
  replaceStart: number;
}

export function parseSubstituteCommand(cmd: string, pos?: number): SubstituteCommand | null {
  const match = cmd.match(/^(.*?)s\/(.*?)\/(.*?)(?:\/([gicI]*))?$/);
  if (!match) {
    return null;
  }
  const [_, range, search, replace, flags = ''] = match;

  const hasVeryMagic = search.startsWith('\\v');
  const searchTerm = hasVeryMagic ? search.substring(2) : search;

  // Calculate field positions
  const searchStart = range.length + 2; // After "s/"
  const searchEnd = searchStart + search.length;
  const replaceStart = searchEnd + 1; // After the second "/"
  const replaceEnd = replaceStart + replace.length;

  // Determine which field has focus
  let isInSearchField = false;
  let isInReplaceField = false;
  let searchPos = undefined;
  let replacePos = undefined;

  if (pos !== undefined) {
    isInSearchField = pos >= searchStart && pos <= searchEnd;
    isInReplaceField = pos >= replaceStart && pos <= replaceEnd;

    if (isInSearchField) {
      searchPos = pos - searchStart;
    } else if (isInReplaceField) {
      replacePos = pos - replaceStart;
    }
  }

  return {
    range,
    search: searchTerm,
    replace,
    flags,
    hasVeryMagic,
    hasGlobal: flags.includes('g'),
    hasIgnoreCase: flags.includes('i'),
    hasCaseSensitive: flags.includes('I'),
    hasConfirm: flags.includes('c'),
    searchPos,
    replacePos,
    isInSearchField,
    isInReplaceField,
    searchStart,
    replaceStart
  };
}
