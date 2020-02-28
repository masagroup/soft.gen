package soft.generator.cpp;

import java.io.IOException;
import java.net.URL;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.Arrays;

import org.eclipse.emf.common.util.BasicMonitor;
import org.eclipse.emf.ecore.EObject;
import org.eclipse.emf.ecore.EPackage;
import org.eclipse.emf.ecore.resource.ResourceSet;

import soft.generator.common.Generator;

/**
 * Entry point of the 'Generate' generation module.
 * 
 * @generated NOT
 */
public class Generate extends Generator {
    /**
     * The name of the module.
     */
    public static final String MODULE_FILE_NAME = "/soft/generator/cpp/generate";

    /**
     * The name of the templates that are to be generated.
     */
    public static final String[] TEMPLATE_NAMES = { "generateModel",
                                                    "generateLibraryCMakeSources",
                                                    "generateLibraryCMakeProject",
                                                    "generateTests",
                                                    "generateTestsCMakeSources",
                                                    "generateTestsCMakeProject" };

    private Generate() {
        super(MODULE_FILE_NAME, TEMPLATE_NAMES);
    }

    public static void main(String[] args) {
        try {
            Generate g = new Generate();
            boolean p = g.parse(args);
            if (p) {
                g.initialize();
                g.generate(new BasicMonitor());
            }
        } catch (Exception e) {
            System.out.println(e.getMessage());
        }
    }

    protected void postGenerate(ResourceSet resourceSet) {
        if (Arrays.stream(getTemplateNames()).anyMatch(t -> t.equals("generateTestsCMakeProject"))) {
            EObject eModel = getModel();
            if (eModel instanceof EPackage) {
                EPackage ePackage = (EPackage) eModel;
                try {
                    URL srcUrl = getClass().getResource("cmake");
                    Path from = Paths.get(srcUrl.toURI()).normalize();
                    Path to = Paths.get(targetFolder.toURI()).resolve(ePackage.getName() + "/tests/cmake").normalize();
                    Files.walk(from).forEach(src -> {
                        final Path dest = to.resolve(from.relativize(src).toString());
                        try {
                            if (Files.isDirectory(src)) {
                                if (Files.notExists(dest)) {
                                    System.out.println("creating directory:" + to);
                                    Files.createDirectories(dest);
                                }
                            } else {
                                System.out.println("extracting file:" + dest);
                                Files.copy(src, dest, StandardCopyOption.REPLACE_EXISTING);
                            }
                        } catch (IOException e) {
                            e.printStackTrace();
                        }
                    });
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }

        }
    }

}
