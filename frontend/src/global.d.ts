type RipgrepResult = [key: string]: Array<main.RipgrepMatch>

interface RipgrepMatch {
  Path: string;
  Row: string;
  Col: string;
  MatchedLine: string;

}
