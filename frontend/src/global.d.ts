type RipgrepResultApi = [key: string]: Array<main.RipgrepMatch>

interface RipgrepResult {
  path : string;
  matches: RipgrepMatch[]
}

interface RipgrepMatch {
  Path: string;
  Row: string;
  Col: string;
  MatchedLine: string;

}
