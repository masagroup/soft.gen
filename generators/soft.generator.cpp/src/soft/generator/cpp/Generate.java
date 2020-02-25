package soft.generator.cpp;

import java.io.InputStream;
import java.util.Arrays;

import org.eclipse.emf.common.util.BasicMonitor;
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
            InputStream s = getClass().getResourceAsStream("cmake/BoostTest.cmake");
            if (s == null)
                System.out.println("not found");
            else
                System.out.println("found");
        }
    }

}
