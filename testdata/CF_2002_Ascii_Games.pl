#!/usr/bin/perl
# Convert the txt file into a CSV input file for cliquetool.
# Original file downloaded from:
#   http://www.phys.utk.edu/sorensen/cfr/cfr/Output/2002/CF_2002_Ascii_Games.txt
# Usage:
#   perl < CF_2002_Ascii_Games.txt > 2002_NCAAF.csv

while ( my $line = <> ) {
  chomp $line;
  my $team1 = substr($line, 32, 30);
  $team1 =~ s/^\s+|\s+$//g;
  my $team2 = substr($line, 63, 30);
  $team2 =~ s/^\s+|\s+$//g;
  print "\"$team1\",\"$team2\"\n";
}
