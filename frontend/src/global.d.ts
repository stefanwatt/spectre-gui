type Modifier = 'c' | 's' | 'a'

interface RipgrepResult {
  path: string;
  matches: RipgrepMatch[]
}

interface RipgrepMatch {
  Path: string;
  Row: string;
  Col: string;
  MatchedLine: string;

}

type RipgrepResultApi = { [key: string]: Array<RipgrepMatch> };
