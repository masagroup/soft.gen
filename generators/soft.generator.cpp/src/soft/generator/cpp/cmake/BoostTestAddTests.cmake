set(prefix "${TEST_PREFIX}")
set(suffix "${TEST_SUFFIX}")
set(extra_args ${TEST_EXTRA_ARGS})
set(properties ${TEST_PROPERTIES})
set(script)
set(suite)
set(tests)

function(add_command NAME)
  set(_args "")
  foreach(_arg ${ARGN})
    if(_arg MATCHES "[^-./:a-zA-Z0-9_]")
      set(_args "${_args} [==[${_arg}]==]")
    else()
      set(_args "${_args} ${_arg}")
    endif()
  endforeach()
  set(script "${script}${NAME}(${_args})\n" PARENT_SCOPE)
endfunction()

# Run test executable to get list of available tests
if(NOT EXISTS "${TEST_EXECUTABLE}")
  message(FATAL_ERROR
    "Specified test executable does not exist.\n"
    "  Path: '${TEST_EXECUTABLE}'"
  )
endif()
execute_process(
  COMMAND ${TEST_EXECUTOR} "${TEST_EXECUTABLE}" --list_content
  TIMEOUT ${TEST_DISCOVERY_TIMEOUT}
  OUTPUT_VARIABLE stdout
  ERROR_VARIABLE stderr
  RESULT_VARIABLE result
)
if(NOT ${result} EQUAL 0)
  string(REPLACE "\n" "\n    " output "${output}")
  message(FATAL_ERROR
    "Error running test executable.\n"
    "  Path: '${TEST_EXECUTABLE}'\n"
    "  Result: ${result}\n"
    "  Output:\n"
    "    ${output}\n"
  )
endif()

if ( NOT "d${stdout}" STREQUAL "d" )
    set( output ${stdout} )
elseif ( NOT "d${stderr}" STREQUAL "d" )
    set( output ${stderr} )
endif()

string(REPLACE "\n" ";" output "${output}")

# Parse output
set( test_set )
set( test_regexp "^( *)([^*]*)([*]+)" )
foreach(line ${output})
    if ( line MATCHES ${test_regexp} )
        
        # Test name
        string(REGEX REPLACE ${test_regexp} "\\1" test_level ${line})
        string(REGEX REPLACE ${test_regexp} "\\2" test_name ${line})
        string(REGEX REPLACE ${test_regexp} "\\3" test_enabled ${line})
        string(LENGTH "${test_level}" test_set_size)
        math(EXPR test_set_size "${test_set_size}/4")
        list(SUBLIST test_set 0 ${test_set_size} test_set)
        list(APPEND  test_set ${test_name})
        list(JOIN test_set "." test_id)
        list(JOIN test_set "/" test_path)
        
        # Add to script
        add_command(add_test "${prefix}${test_id}${suffix}" 
                             ${TEST_EXECUTOR} 
                             "${TEST_EXECUTABLE}" 
                             "--run_test=${test_path}" 
                             "--catch_system_error=yes" ${extra_args} 
                    )
                    
        if(NOT test_enabled MATCHES "^\\*")
            add_command(set_tests_properties "${prefix}${test_id}${suffix}" 
                                              PROPERTIES DISABLED TRUE)
        endif()

        add_command(set_tests_properties "${prefix}${test_id}${suffix}" 
                                          PROPERTIES WORKING_DIRECTORY "${TEST_WORKING_DIR}" 
                                          ${properties} )
        
        list(APPEND tests "${prefix}${test_id}${suffix}")
    endif()
endforeach()

# Create a list of all discovered tests, which users may use to e.g. set
# properties on the tests
add_command(set ${TEST_LIST} ${tests})

# Write CTest script
file(WRITE "${CTEST_FILE}" "${script}")
