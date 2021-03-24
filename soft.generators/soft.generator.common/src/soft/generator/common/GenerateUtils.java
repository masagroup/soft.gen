/*
 * This file is part of soft.generators, a project for code 
 * generation of an ecore model
 * 
 * Copyright(c) 2021 MASA Group
 * 
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package soft.generator.common;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Set;

import org.eclipse.emf.ecore.EClass;
import org.eclipse.emf.ecore.EClassifier;
import org.eclipse.emf.ecore.EPackage;

import com.google.common.base.CaseFormat;

public class GenerateUtils {

    public String lowerCamelToUpperUnderscore(String s) {
        return CaseFormat.LOWER_CAMEL.to(CaseFormat.UPPER_UNDERSCORE, s);
    }

    public String upperCamelToUpperUnderscore(String s) {
        return CaseFormat.UPPER_CAMEL.to(CaseFormat.UPPER_UNDERSCORE, s);
    }

    /**
     * Formats a string by splitting it into words separated by underscores and/or
     * mixed-casing and then recombining them using the specified separator.
     */
    public String splitAndCombineWords(String s, String separator) {
        List<String> parsedName = new ArrayList<String>();
        if (s.length() != 0)
            parsedName.addAll(splitWords(s, '_'));

        StringBuilder result = new StringBuilder();
        for (Iterator<String> nameIter = parsedName.iterator(); nameIter.hasNext();) {
            String nameComponent = nameIter.next();
            result.append(nameComponent);

            if (nameIter.hasNext() && nameComponent.length() > 1)
                result.append(separator);
        }
        return result.toString();
    }

    /**
     * This method breaks s into words delimited by separator and/or mixed-case
     * naming.
     */
    private List<String> splitWords(String s, char separator) {
        List<String> result = new ArrayList<String>();
        if (s != null) {
            StringBuilder currentWord = new StringBuilder();
            boolean lastIsLower = false;
            for (int index = 0, length = s.length(); index < length; ++index) {
                char curChar = s.charAt(index);
                if (Character.isUpperCase(curChar) || (!lastIsLower && Character.isDigit(curChar))
                        || curChar == separator) {
                    if (lastIsLower && currentWord.length() > 1 || curChar == separator && currentWord.length() > 0) {
                        result.add(currentWord.toString());
                        currentWord = new StringBuilder();
                    }
                    lastIsLower = false;
                } else {
                    if (!lastIsLower) {
                        int currentWordLength = currentWord.length();
                        if (currentWordLength > 1) {
                            char lastChar = currentWord.charAt(--currentWordLength);
                            currentWord.setLength(currentWordLength);
                            result.add(currentWord.toString());
                            currentWord = new StringBuilder();
                            currentWord.append(lastChar);
                        }
                    }
                    lastIsLower = true;
                }

                if (curChar != separator) {
                    currentWord.append(curChar);
                }
            }

            result.add(currentWord.toString());
        }
        return result;
    }

    public List<EClass> getOrderedClasses(EPackage ePackage) {
        List<EClass> result = new ArrayList<EClass>();
        Set<EClass> resultSet = new HashSet<EClass>();
        for (EClassifier eClassifier : ePackage.getEClassifiers()) {
            if (eClassifier instanceof EClass) {
                EClass eClass = (EClass) eClassifier;
                List<EClass> extendChain = new LinkedList<>();
                Set<EClass> visited = new HashSet<EClass>();
                while (eClass != null && visited.add(eClass)) {
                    if (ePackage == eClass.getEPackage() && resultSet.add(eClass))
                        extendChain.add(0, eClass);

                    eClass = getSuperType(eClass);
                }
                result.addAll(extendChain);
            }
        }
        ;
        return result;
    }

    private EClass getSuperType(EClass eClass) {
        return eClass.getESuperTypes().stream().findFirst().orElse(null);
    }

    public Map<Object, Object> mapNew() {
        return new HashMap<>();
    }

    public Map<Object, Object> mapPut(Map<Object, Object> map, Object key, Object value) {
        map.put(key, value);
        return map;
    }

    public Map<Object, Object> mapPutAll(Map<Object, Object> map, Map<Object, Object> o) {
        map.putAll(o);
        return map;
    }

    public Map<Object, Object> mapRemove(Map<Object, Object> map, Object key) {
        map.remove(key);
        return map;
    }

    public Object mapGet(Map<Object, Object> map, Object key) {
        return map.get(key);
    }

    public Object mapGetOrDefault(Map<Object, Object> map, Object key, Object defaultValue) {
        return map.getOrDefault(key, defaultValue);
    }

    public Set<Object> mapKeys(Map<Object, Object> map) {
        return map.keySet();
    }

    public boolean mapContainsKey(Map<Object, Object> map, Object key) {
        return map.containsKey(key);
    }

    @SuppressWarnings("unchecked")
    public List<Object> asSequence(Object obj) {
        return (List<Object>) obj;
    }
}
