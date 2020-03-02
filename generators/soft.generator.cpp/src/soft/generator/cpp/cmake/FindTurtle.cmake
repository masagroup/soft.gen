# - Find Turtle
# Find the Turtle (http://turtle.sourceforge.net/) includes
#
# This module defines the following :prop_tgt:`IMPORTED` targets:
#    Turtle::Turtle
# This module defines
#  Turtle_INCLUDE_DIR, Turtle include directory, where to find turtle/mock.hpp, etc.
#  Turtle_FOUND, If false, do not try to use Turtle.
#
# This module reads hints about search locations from variables:
#  TURTLE_ROOT          - Preferred installation directory
#  TURTLE_INCLUDEDIR    - Preferred include directory

if(NOT TURTLE_ROOT)
    set(TURTLE_ROOT "" CACHE PATH "Installation directory of Turtle")
endif()

find_path(Turtle_INCLUDE_DIR
    NAMES turtle/mock.hpp turtle/sequence.hpp
    PATHS ${TURTLE_ROOT} ${TURTLE_INCLUDEDIR}
    PATH_SUFFIXES include
    )
mark_as_advanced(Turtle_INCLUDE_DIR)

# handle the QUIETLY and REQUIRED arguments and set Turtle_FOUND to TRUE if
# all listed variables are TRUE
include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(Turtle DEFAULT_MSG Turtle_INCLUDE_DIR)

if(Turtle_FOUND AND NOT TARGET Turtle::Turtle)
    add_library(Turtle::Turtle INTERFACE IMPORTED)
    if(Turtle_INCLUDE_DIR)
      set_target_properties(Turtle::Turtle PROPERTIES INTERFACE_INCLUDE_DIRECTORIES "${Turtle_INCLUDE_DIR}")
    endif()
endif()
