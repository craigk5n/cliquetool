# This Makefile will use the sample data to process some data and then
# create reports.

CLIQUETOOL = ./cliquetool

ALL: doc/2018_NCAAB_report.md \
	doc/2002_NCAAF_report.md

# Find all cliques in the 2018 college basketball schedule data.
# Then create a report on the results in markdown format.
doc/2018_NCAAB_report.md: CLIQUETOOL testdata/2018_NCAAB.csv
	$(CLIQUETOOL) -infile testdata/2018_NCAAB.csv \
        	-passes 1 -minsize 8 -find -sort -rename\
		-writecsv 2018_NCAAB-cliques.csv
	$(CLIQUETOOL) -infile testdata/2018_NCAAB.csv \
		 -loadgroupcsv 2018_NCAAB-cliques.csv -verify doc/2018_NCAAB_report.md

# Find all cliques in the 2002 college football schedule data.
# Use that to build a list of dense groups and only consider groups of size 8 or larger.
# Then create a report on the results in markdown format.
doc/2002_NCAAF_report.md: CLIQUETOOL testdata/2002_NCAAF.csv
	$(CLIQUETOOL) -i testdata/2002_NCAAF.csv -minsize 5 -find -sort -rename \
		-writecsv 2002_NCAAF-cliques.csv -param testdata/2002_NCAAF-params.txt \
		-passes 3 -build -mergeparams 0.7 -merge -prune -prefix NCAAF_Conf_ \
		-sort -rename \
		-minsize 8 -writecsv 2002_NCAAF-build-groups.csv
	$(CLIQUETOOL) -i testdata/2002_NCAAF.csv \
		-loadgroupcsv 2002_NCAAF-build-groups.csv -verify doc/2002_NCAA_report.md

# Generate the 2002 CF CSV file from a downloaded text file using a simple perl script
#testdata/2002_NCAAF.csv: testdata/CF_2002_Ascii_Games.txt
#	perl testdata/CF_2002_Ascii_Games.pl < testdata/CF_2002_Ascii_Games.txt \
#		> testdata/2002_NCAAF.csv


