package soft.acceleo;

import org.eclipse.acceleo.internal.parser.compiler.IAcceleoParserURIHandler;
import org.eclipse.emf.common.util.URI;

public class AcceleoURIHandler implements IAcceleoParserURIHandler {

	@Override
	public URI transform(URI uri) {
		URI newURI = uri;
		if (newURI.toString().startsWith("jar:file:")) {
			int indexOf = newURI.toString().indexOf(".jar!/");
			if (indexOf != -1) {
				String name = newURI.toString();
				name = name.substring(0, indexOf);
				name = name.substring("jar:file:".length() + 1);
				if (name.endsWith("-SNAPSHOT")) {
					name = name.substring(0, name.length() - "-SNAPSHOT".length());
				}

				name = name.substring(0, name.lastIndexOf("-"));
				if (name.contains("/")) {
					name = name.substring(name.lastIndexOf("/"));
					name = name + "/";
				}
				name = "platform:/plugin" + name + newURI.toString().substring(indexOf + ".jar!/".length());
				newURI = URI.createURI(name);
			}
		}
		return newURI;
	}

}
