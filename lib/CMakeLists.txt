cmake_minimum_required(VERSION 3.10)
project(goopenjpeg C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

set(OPENJPEG_DIR ${CMAKE_CURRENT_SOURCE_DIR}/openjpeg)
set(OPENJPEG_BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/openjpeg-build)
set(INTERFACE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/interface)

set(BUILD_CODEC OFF CACHE BOOL "" FORCE)
set(BUILD_JPIP OFF CACHE BOOL "" FORCE)
set(BUILD_VIEWER OFF CACHE BOOL "" FORCE)
set(BUILD_JAVA OFF CACHE BOOL "" FORCE)
set(BUILD_UNIT_TESTS OFF CACHE BOOL "" FORCE)
set(BUILD_STATIC_LIBS ON CACHE BOOL "" FORCE)
set(BUILD_SHARED_LIBS OFF CACHE BOOL "" FORCE)

add_subdirectory(${OPENJPEG_DIR} ${OPENJPEG_BINARY_DIR})

file(GLOB LCMS2_SOURCES ${OPENJPEG_DIR}/thirdparty/liblcms2/src/*.c)

add_library(goopenjpeg SHARED
    ${INTERFACE_DIR}/memstream.c
    ${INTERFACE_DIR}/j2k_decode.c
    ${INTERFACE_DIR}/color.c
    ${LCMS2_SOURCES}
    ${CMAKE_CURRENT_SOURCE_DIR}/capi/goopenjpeg.c
)

target_include_directories(goopenjpeg PRIVATE
    ${OPENJPEG_DIR}/src/lib/openjp2
    ${OPENJPEG_BINARY_DIR}/src/lib/openjp2
    ${OPENJPEG_DIR}/src/bin/common
    ${OPENJPEG_DIR}/thirdparty/liblcms2/include
    ${INTERFACE_DIR}
)

add_dependencies(goopenjpeg openjp2)
target_link_libraries(goopenjpeg PRIVATE openjp2)

set_target_properties(goopenjpeg PROPERTIES
    LIBRARY_OUTPUT_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}"
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}"
    ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}"
)

if(MSVC)
    set_target_properties(goopenjpeg PROPERTIES
        WINDOWS_EXPORT_ALL_SYMBOLS ON
    )
    foreach(flag CMAKE_C_FLAGS CMAKE_C_FLAGS_DEBUG CMAKE_C_FLAGS_RELEASE)
        if(${flag} MATCHES "/MD")
            string(REGEX REPLACE "/MD" "/MT" ${flag} "${${flag}}")
        endif()
    endforeach()
endif()

if(CMAKE_COMPILER_IS_GNUCC OR CMAKE_C_COMPILER_ID MATCHES "Clang")
    set_target_properties(goopenjpeg PROPERTIES
        C_VISIBILITY_PRESET hidden
        VISIBILITY_INLINES_HIDDEN ON
    )
endif()

if(WIN32)
    set_target_properties(goopenjpeg PROPERTIES PREFIX "" SUFFIX ".dll")
elseif(APPLE)
    set_target_properties(goopenjpeg PROPERTIES PREFIX "lib" SUFFIX ".dylib")
else()
    set_target_properties(goopenjpeg PROPERTIES PREFIX "lib" SUFFIX ".so")
endif()
