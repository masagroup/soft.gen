package soft.generator.go;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.function.Function;
import java.util.stream.Collectors;

import com.google.common.collect.Lists;

public class GeneratePath {

    public Map<String, String> getPackageImports(Set<String> packages) {
        Map<String, String[]> splitted = new HashMap<String, String[]>();
        Map<String, List<String>> aliases = new HashMap<String, List<String>>();
        int[] index = new int[1];

        // for each remaining paths
        List<String> paths = new ArrayList<String>(packages);
        while (!paths.isEmpty()) {
            // compute all package names
            Map<String, String> packagesNames = packages.stream()
                                                        .collect(Collectors.toMap(Function.identity(), pack -> {
                                                            int i = index[0];
                                                            String[] path = splitted.computeIfAbsent(pack,
                                                                                                     p -> reverseArray(p.split("/")));
                                                            return path.length > i ? path[i] : null;
                                                        }));

            // compute path names
            List<String> names = paths.stream().map(p -> packagesNames.get(p)).toList();

            // for each name, check if there is another one which is equal
            for (int i = names.size() - 1; i >= 0; i--) {
                String name = names.get(i);
                String path = paths.get(i);
                if (name == null || name.length() == 0) {
                    paths.remove(i);
                } else {
                    boolean other = false;
                    for (Map.Entry<String, String> entry : packagesNames.entrySet()) {
                        String p = entry.getKey();
                        String n = entry.getValue();
                        if (name.equals(n) && !p.equals(path)) {
                            other = true;
                            break;
                        }
                    }

                    List<String> alias = aliases.computeIfAbsent(path, t -> new ArrayList<String>());
                    alias.add(name);
                    if (!other) {
                        paths.remove(i);
                    }
                }
            }
            index[0]++;
        }

        // build name for each package
        Map<String, String> result = new HashMap<String, String>();
        for (Map.Entry<String, List<String>> entry : aliases.entrySet()) {
            String name = String.join("_", Lists.reverse(entry.getValue()));
            result.put(entry.getKey(), name);
        }

        return result;
    }

    private String[] reverseArray(String[] a) {
        for (int i = 0; i < a.length / 2; i++) {
            String temp = a[i];
            a[i] = a[a.length - i - 1];
            a[a.length - i - 1] = temp;
        }
        return a;
    }
}
