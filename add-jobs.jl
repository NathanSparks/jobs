# Add jobs to swif workflow and submit it to farm
# takes 1 argument, a config-file
if length(ARGS) != 1 throw("requires 1 argument: the config. file.") end
cdata = readdlm(ARGS[1],ASCIIString)
d = Dict{ASCIIString,ASCIIString}()
inputs = Array(Array{ASCIIString,1},0)
outputs = Array(Array{ASCIIString,1},0)
for i=1:size(cdata,1)
    if cdata[i,1] != "output" && cdata[i,1] != "input"
        d[cdata[i,1]] = cdata[i,2]
    elseif cdata[i,1] == "input"
        push!(inputs,[cdata[i,2],cdata[i,3]])
    elseif cdata[i,1] == "output"
        push!(outputs,[cdata[i,2],cdata[i,3]])
    end
end
if d["flag"] != "test" && d["flag"] != "submit" throw("not a valid 'flag'; set to 'test' to see what will be submitted for a few jobs, or 'submit' to submit all jobs to farm") end
if length(outputs) == 0 throw("no valid job output files; check for typos in config. file, '$(ARGS[1])'.") end
if d["flag"] == "test" println(d); println(outputs) end
USE_DATADIR = false
if haskey(d,"copy-inputdir") USE_DATADIR = true end
DATADIR = d["inputdir"]
PREFIX = d["inputfile-prefix"]
SUFFIX = d["inputfile-suffix"]
MINRUNNO = parse(Int,d["runno-min"])
MAXRUNNO = parse(Int,d["runno-max"])
MINFILENO = parse(Int,d["fileno-min"])
MAXFILENO = parse(Int,d["fileno-max"])
SH = endswith(d["shell"],"/tcsh") ? `$(d["shell"]) -f` : d["shell"]
ENVIRON = joinpath(pwd(),d["environ"])
SCRIPT = joinpath(pwd(),d["script"])
USE_CONFIG = true
if haskey(d,"hdconfig") CONFIG = joinpath(pwd(),d["hdconfig"])
else CONFIG = ""; USE_CONFIG = false end
info("making filelist_$(d["runno-min"])-$(d["runno-max"]).txt")
OUTFILE = open("filelist_$(d["runno-min"])-$(d["runno-max"]).txt","w")
RUNS = Int[]
if !contains(DATADIR,"[runno]") MAXRUNNO = MINRUNNO end
if haskey(d,"runlist")
    if d["runlist"] == "rcdb"
        ENV["PATH"] = string("/apps/python/PRO/bin:",ENV["PATH"])
        if haskey(ENV,"LD_LIBRARY_PATH") ENV["LD_LIBRARY_PATH"] = string("/apps/python/PRO/lib:",ENV["LD_LIBRARY_PATH"])
        else ENV["LD_LIBRARY_PATH"] = "/apps/python/PRO/lib" end
        run(`python mkrunlist.py $(d["runno-min"]) $(d["runno-max"])`)
        RUN_LIST = open("runlist_$(d["runno-min"])-$(d["runno-max"]).txt")
    else
        RUN_LIST = open(d["runlist"])
    end
    for line in eachline(RUN_LIST) push!(RUNS,parse(Int,chomp(line))) end
    close(RUN_LIST)
else
    RUNS = collect(MINRUNNO:MAXRUNNO)
end
for RUNNO=MINRUNNO:MAXRUNNO
    if !(RUNNO in RUNS) continue end
    DIR = contains(DATADIR,"[runno]") ? replace(DATADIR,"[runno]",string("0",RUNNO)) : DATADIR
    if ispath(DIR)
        fileno = 0
        for FILE in readdir(DIR)
            if contains(FILE,PREFIX) && contains(FILE,SUFFIX)
                if fileno >= MINFILENO && fileno <= MAXFILENO println(OUTFILE,DIR,"/",chomp(FILE)) end
                fileno += 1
            end
        end
    end
end
close(OUTFILE)
info("adding jobs to '$(d["workflow"])' workflow")
Nruns = 0; Nfiles = 0
PREV_RUNNO = -1
FILE_LIST = open("filelist_$(d["runno-min"])-$(d["runno-max"]).txt")
lPREFIX = length(PREFIX); lSUFFIX = length(SUFFIX)
INPUTDIR_PREFIX = "file:"
if startswith(DATADIR,"/mss/") INPUTDIR_PREFIX = "mss:" end
for INPUT in eachline(FILE_LIST)
    INPUT = chomp(INPUT)
    # get runno from filename
    LOC_INPUT = basename(INPUT)
    i1 = lPREFIX + 1; i2 = length(LOC_INPUT) - lSUFFIX
    FILE_ID = LOC_INPUT[i1:i2]
    #RUNNO = parse(Int,FILE_ID[1:6])
    RUNNO_str = split(FILE_ID,"_")[1]
    RUNNO = parse(Int,RUNNO_str)
    #FILENO = parse(Int,split(FILE_ID,"_")[2])
    # count number of runs and files
    if RUNNO != PREV_RUNNO Nruns += 1 end
    PREV_RUNNO = RUNNO
    Nfiles += 1
    if USE_DATADIR  FILE_ID = RUNNO_str end
    NAME = replace(d["name"],"[file-id]",FILE_ID)
    STDOUT = replace(d["stdout"],"[file-id]",FILE_ID)
    STDERR = replace(d["stderr"],"[file-id]",FILE_ID)
    INPUT = string(INPUTDIR_PREFIX,INPUT)
    # allow for multiple inputs
    if !USE_DATADIR INPUT_CMD = `-input $LOC_INPUT $INPUT`
    else INPUT_CMD = `` end
    for input in inputs
        input = [replace(input[1],"[file-id]",FILE_ID),replace(input[2],"[file-id]",FILE_ID)]
        INPUT_CMD = `$INPUT_CMD -input $(input[1]) $(input[2])`
    end
    # allow for multiple outputs
    OUTPUT_CMD = ``
    for output in outputs
        output = [replace(output[1],"[file-id]",FILE_ID),replace(output[2],"[file-id]",FILE_ID)]
        OUTPUT_CMD = `$OUTPUT_CMD -output $(output[1]) $(output[2])`
    end
    if !USE_CONFIG CONFIG = RUNNO_str end # set to runno if not set in config. file
    if USE_DATADIR  LOC_INPUT = replace(DATADIR,"[runno]",RUNNO_str) end
    # add job to workflow
    cmd = `swif add-job -create -workflow $(d["workflow"]) -project $(d["project"]) -track $(d["track"]) -cores $(d["cores"]) -disk $(d["disk"]) -ram $(d["ram"]) -time $(d["time"]) -os $(d["os"])
    $INPUT_CMD $OUTPUT_CMD
    -stdout $STDOUT
    -stderr $STDERR
    -name $NAME $SH $SCRIPT $ENVIRON $LOC_INPUT $CONFIG`
    if d["flag"] == "test" println(cmd) end
    if d["flag"] == "submit" run(cmd) end
end
close(FILE_LIST)
# submit workflow to farm
info(string(Nfiles," jobs added, 1 job/file"))
if Nruns == 0 info("no jobs to submit"); exit(0) end
info(string(Nruns," runs to analyze, average of ",round(Int,Nfiles/Nruns)," files/run"))
if d["flag"] == "submit"
    info("submitting '$(d["workflow"])' workflow")
    run(`swif run -workflow $(d["workflow"])`)
end
